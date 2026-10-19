package bakery

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bakery-scheduler/internal/entity/bakery"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const dayFmt = "2006-01-02"

func fmtDay(t time.Time) string { return t.Format(dayFmt) }

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now().In(s.loc)
	}
	return time.Now().In(s.loc)
}

// Loc zona waktu aplikasi (dipakai handler untuk membaca tanggal).
func (s Service) Loc() *time.Location { return s.loc }

func (s Service) parseDay(v string) (time.Time, error) {
	return time.ParseInLocation(dayFmt, v, s.loc)
}

// retryDeadlock mengulang operasi (maksimal 4 kali) bila MySQL membatalkan
// transaksi karena deadlock (1213) atau lock wait timeout (1205).
func retryDeadlock[T any](fn func() (T, error)) (T, error) {
	var out T
	var err error
	for attempt := 0; attempt < 4; attempt++ {
		out, err = fn()
		var me *mysql.MySQLError
		if err == nil || !errors.As(err, &me) || (me.Number != 1213 && me.Number != 1205) {
			return out, err
		}
		time.Sleep(time.Duration(attempt+1) * 15 * time.Millisecond)
	}
	return out, err
}

// audit mencatat jejak perubahan. before/after di-encode ke JSON.
func (s Service) audit(ctx context.Context, tx *sqlx.Tx, orderID int64, actor, aksi string, before, after interface{}, alasan string) error {
	enc := func(v interface{}) *string {
		if v == nil {
			return nil
		}
		b, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		str := string(b)
		return &str
	}
	return s.bakery.InsertAudit(ctx, tx, orderID, aksi, enc(before), enc(after), alasan, actor)
}

func (s Service) notify(ctx context.Context, tx *sqlx.Tx, orderID int64, msg string) error {
	return s.bakery.InsertNotification(ctx, tx, orderID, msg)
}

// loadSettings membaca parameter produksi + bobot AHP aktif.
func (s Service) loadSettings(ctx context.Context, tx *sqlx.Tx) (bakery.Settings, error) {
	set := DefaultSettings()
	m, err := s.bakery.GetSettings(ctx, tx)
	if err != nil {
		return set, err
	}
	for k, f := range m {
		switch k {
		case "kapasitas_harian":
			set.CapacityPerDay = int(f)
		case "lembur_maks":
			set.OvertimeMax = int(f)
		case "jendela_jam":
			set.WindowHours = f
		case "batas_pagi_jam":
			set.MorningCutoffHour = int(f)
		case "lama_menunggu_maks_jam":
			set.MaxWaitHours = f
		case "pembulatan_jam":
			set.RoundHours = int(f)
		case "ambang_selisih_skor":
			set.MinScoreGap = f
		case "maks_tawaran_jam":
			set.MaxOffers = int(f)
		case "batas_respons_menit":
			set.HoldMinutes = int(f)
		case "jam_produksi_selesai":
			set.ProductionEndHour = int(f)
		case "menit_packing":
			set.PackingMinutes = int(f)
		case "menit_tempuh":
			set.TravelMinutes = int(f)
		case "jam_kirim_terakhir":
			set.LastDeliveryHour = int(f)
		case "jam_peringatan_sebelum_selesai":
			set.WarnHours = int(f)
		}
	}
	w, err := s.bakery.GetActiveWeights(ctx, tx)
	if err == nil {
		set.Weights = w
	} else if !errors.Is(err, sql.ErrNoRows) {
		return set, err
	}
	return set, nil
}

// getDay membuat baris kapasitas hari itu bila belum ada, MENGUNCINYA
// (SELECT ... FOR UPDATE), lalu menghitung kapasitas terpakai.
func (s Service) getDay(ctx context.Context, tx *sqlx.Tx, day time.Time, set bakery.Settings, now time.Time) (bakery.DayState, error) {
	st := bakery.DayState{Day: day}
	ds := fmtDay(day)
	var err error
	st.Base, st.OvertimeMax, st.Locked, err = s.bakery.LockDay(ctx, tx, ds)
	if errors.Is(err, sql.ErrNoRows) { // jarang: baris biasanya sudah dibuat oleh ensureDays
		if err := s.bakery.InsertDayIgnore(ctx, tx, ds, set.CapacityPerDay, set.OvertimeMax); err != nil {
			return st, err
		}
		st.Base, st.OvertimeMax, st.Locked, err = s.bakery.LockDay(ctx, tx, ds)
	}
	if err != nil {
		return st, err
	}
	st.Used, err = s.bakery.SumUsed(ctx, tx, ds, now)
	return st, err
}

// ensureDays membuat baris kapasitas untuk rentang hari DI LUAR transaksi
// (autocommit). Membuat baris di dalam transaksi yang berebut memicu deadlock.
func (s Service) ensureDays(ctx context.Context, from, to time.Time, set bakery.Settings) error {
	for d := DayOf(from); !d.After(to); d = AddDays(d, 1) {
		if err := s.bakery.InsertDayIgnore(ctx, nil, fmtDay(d), set.CapacityPerDay, set.OvertimeMax); err != nil {
			return err
		}
	}
	return nil
}

var bulanID = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

// idDate: "2026-10-14" -> "14 Okt" (untuk pesan galat).
func idDate(v string) string {
	var y, m, d int
	if _, err := fmt.Sscanf(v, "%d-%d-%d", &y, &m, &d); err != nil || m < 1 || m > 12 {
		return v
	}
	return fmt.Sprintf("%d %s", d, bulanID[m-1])
}

// statusLabel menurunkan label tampilan untuk penjual dari status internal.
//
//	baru / menunggu_negosiasi -> "Baru"
//	terjadwal                 -> "Terjadwal 14 Okt", atau "Diproduksi hari ini" bila hari produksinya hari ini
//	diproduksi                -> "Diproduksi hari ini"
//	at_risk                   -> "Berisiko"
//	selesai / terlambat / batal tetap apa adanya
func statusLabel(status string, prodDay *string, today string) string {
	switch status {
	case "baru", "menunggu_negosiasi":
		return "Baru"
	case "diproduksi":
		return "Diproduksi hari ini"
	case "at_risk":
		return "Berisiko"
	case "terjadwal":
		if prodDay == nil {
			return "Terjadwal"
		}
		if *prodDay == today {
			return "Diproduksi hari ini"
		}
		var y, m, d int
		if _, err := fmt.Sscanf(*prodDay, "%d-%d-%d", &y, &m, &d); err != nil || m < 1 || m > 12 {
			return "Terjadwal"
		}
		return fmt.Sprintf("Terjadwal %d %s", d, bulanID[m-1])
	case "selesai":
		return "Selesai"
	case "terlambat":
		return "Terlambat"
	case "batal":
		return "Batal"
	}
	return status
}

// applyLabels mengisi status_label, needs_negotiation, dan can_start.
func applyLabels(list []bakery.OrderRow, today string) []bakery.OrderRow {
	for i := range list {
		o := &list[i]
		o.NeedsNego = o.Status == "menunggu_negosiasi"
		o.StatusLabel = statusLabel(o.Status, o.ProductionDay, today)
		o.CanStart = (o.Status == "terjadwal" || o.Status == "at_risk") && o.ProductionDay != nil && *o.ProductionDay <= today
	}
	return list
}
