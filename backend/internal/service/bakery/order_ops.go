package bakery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"bakery-scheduler/internal/entity/bakery"
)

// Transition mengubah status order: start, complete, cancel.
func (s Service) Transition(ctx context.Context, orderID int64, action, actor string) error {
	now := s.now()
	tx, err := s.bakery.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	status, err := s.bakery.GetOrderStatusLock(ctx, tx, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return bakery.ErrNotFound
	}
	if err != nil {
		return err
	}
	var to string
	switch action {
	case "start":
		if status != "terjadwal" && status != "at_risk" {
			return fmt.Errorf("%w: hanya order terjadwal yang bisa mulai diproduksi", bakery.ErrState)
		}
		// Produksi tidak boleh dimulai sebelum hari produksi yang dijadwalkan.
		pd, err := s.bakery.GetOrderProductionDay(ctx, tx, orderID)
		if err != nil {
			return err
		}
		if !pd.Valid || pd.String > fmtDay(DayOf(now)) {
			return fmt.Errorf("%w: produksi baru bisa dimulai pada hari produksi (%s)", bakery.ErrState, idDate(pd.String))
		}
		// Jam operasional produksi: mulai ProductionStartHour (06.00).
		set, err := s.loadSettings(ctx, tx)
		if err != nil {
			return err
		}
		if now.Hour() < set.ProductionStartHour {
			return fmt.Errorf("%w: produksi baru bisa dimulai pukul %02d.00", bakery.ErrState, set.ProductionStartHour)
		}
		to = "diproduksi"
	case "complete":
		if status != "diproduksi" && status != "terlambat" {
			return fmt.Errorf("%w: order belum diproduksi", bakery.ErrState)
		}
		to = "selesai"
	case "cancel":
		if status == "selesai" || status == "batal" {
			return fmt.Errorf("%w: order sudah %s", bakery.ErrState, status)
		}
		to = "batal"
	default:
		return fmt.Errorf("%w: aksi tidak dikenal", bakery.ErrValidation)
	}
	if err := s.bakery.UpdateOrderStatus(ctx, tx, orderID, to); err != nil {
		return err
	}
	if to == "selesai" {
		if err := s.bakery.UpdateOrderDelivered(ctx, tx, orderID, now); err != nil {
			return err
		}
	}
	if to == "batal" { // lepaskan kapasitas dan tawaran yang masih menggantung
		if err := s.bakery.DeleteReservationsByOrder(ctx, tx, orderID); err != nil {
			return err
		}
		if err := s.bakery.ExpireOpenOffersByOrder(ctx, tx, orderID); err != nil {
			return err
		}
	}
	if err := s.audit(ctx, tx, orderID, actor, "status_"+to,
		map[string]string{"status": status}, map[string]string{"status": to}, ""); err != nil {
		return err
	}
	return tx.Commit()
}

// RunHourly: lepas tahanan kedaluwarsa, hitung ulang skor, tandai terlambat,
// kunci jadwal hari berikutnya (frozen window) setelah jam produksi selesai.
func (s Service) RunHourly(ctx context.Context) error {
	now := s.now()
	tx, err := s.bakery.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	set, err := s.loadSettings(ctx, tx)
	if err != nil {
		return err
	}

	// 1. tahanan kedaluwarsa
	exps, err := s.bakery.GetExpiredOffers(ctx, tx, now)
	if err != nil {
		return err
	}
	for _, e := range exps {
		if err := s.bakery.DeleteReservation(ctx, tx, e.ResID); err != nil {
			return err
		}
		if err := s.bakery.UpdateOfferAnswered(ctx, tx, e.ID, "kedaluwarsa", now); err != nil {
			return err
		}
		if err := s.audit(ctx, tx, e.OrderID, "sistem", "negosiasi_kedaluwarsa", nil, nil, "Batas waktu respons lewat"); err != nil {
			return err
		}
	}

	// 1b. order Baru yang sudah masuk jendela perencanaan menjadi Terjadwal
	win := now.Add(time.Duration(set.WindowHours * float64(time.Hour)))
	promote, err := s.bakery.GetNewOrdersInWindow(ctx, tx, win)
	if err != nil {
		return err
	}
	for _, id := range promote {
		if err := s.bakery.UpdateOrderStatus(ctx, tx, id, "terjadwal"); err != nil {
			return err
		}
		if err := s.audit(ctx, tx, id, "sistem", "masuk_jendela", nil, nil, "Masuk jendela perencanaan 72 jam, status menjadi Terjadwal"); err != nil {
			return err
		}
	}

	// 2. tandai terlambat
	if err := s.bakery.MarkLateOrders(ctx, tx, now); err != nil {
		return err
	}

	// 3. hitung ulang skor
	rows, err := s.bakery.GetOrdersForRescore(ctx, tx)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if err := s.bakery.UpdateOrderScore(ctx, tx, r.ID, ComputeScore(r.Input, now, set), now); err != nil {
			return err
		}
	}

	// 4. frozen window
	if now.Hour() >= set.ProductionEndHour {
		tomorrow := AddDays(DayOf(now), 1)
		if err := s.bakery.LockDaysUntil(ctx, tx, fmtDay(tomorrow)); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return s.rebalance(ctx)
}

func (s Service) rebalance(ctx context.Context) error {
	_, err := retryDeadlock(func() (struct{}, error) { return struct{}{}, s.rebalanceOnce(ctx) })
	return err
}

// rebalanceOnce dijalankan tiap jam setelah skor dihitung ulang.
//
//  1. Menggeser ke hari H: order Baru/Terjadwal (tidak harus sudah masuk jendela
//     72 jam) berjam kirim >= jam siap (11.00) yang masih diproduksi H-1
//     dipindah ke hari kirimnya bila kapasitas normal hari H masih cukup, tanpa
//     menunggu H-1 penuh. H-1 jadi longgar untuk order yang wajib H-1.
//     Hari yang dikunci (frozen) dan order yang sedang diproduksi tidak disentuh.
//     Order tidak pernah dimajukan kembali ke H-1 agar tidak bolak-balik.
//  2. Menandai At risk: order terjadwal yang hari produksinya hari ini (atau
//     sudah lewat) tetapi belum mulai diproduksi saat jam produksi harian
//     hampir selesai (WarnHours jam sebelumnya).
func (s Service) rebalanceOnce(ctx context.Context) error {
	now := s.now()
	set, err := s.loadSettings(ctx, nil)
	if err != nil {
		return err
	}
	today := DayOf(now)
	if err := s.ensureDays(ctx, today, AddDays(today, 4), set); err != nil {
		return err
	}
	tx, err := s.bakery.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	all, err := s.bakery.GetRebalanceItems(ctx, tx, fmtDay(today))
	if err != nil {
		return err
	}
	var items []bakery.RebalanceItem
	for _, it := range all {
		if it.Deadline, err = s.parseDay(it.DeadlineStr); err != nil {
			return err
		}
		if it.Cur, err = s.parseDay(it.CurStr); err != nil {
			return err
		}
		// hanya order yang masih punya hari cadangan (hari H) setelah hari produksinya
		if it.Deadline.After(it.Cur) {
			it.Score = ComputeScore(bakery.ScoreInput{Qty: it.Qty, DeliverAt: it.Deliver, OrderedAt: it.Order}, now, set).Total
			items = append(items, it)
		}
	}
	// Hari dikunci urut naik (hindari deadlock); dalam satu hari skor terendah dulu.
	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].Cur.Equal(items[j].Cur) {
			return items[i].Cur.Before(items[j].Cur)
		}
		return items[i].Score < items[j].Score
	})

	for _, it := range items {
		from, err := s.getDay(ctx, tx, it.Cur, set, now)
		if err != nil {
			return err
		}
		to, err := s.getDay(ctx, tx, it.Deadline, set, now)
		if err != nil {
			return err
		}
		if from.Locked || to.Locked || to.Free() < it.Qty {
			continue
		}
		if err := s.bakery.UpdateReservationDay(ctx, tx, it.ResID, fmtDay(it.Deadline)); err != nil {
			return err
		}
		if err := s.bakery.UpdateOrderProductionDay(ctx, tx, it.OrderID, fmtDay(it.Deadline)); err != nil {
			return err
		}
		if err := s.audit(ctx, tx, it.OrderID, "sistem", "digeser",
			map[string]string{"hari_produksi": fmtDay(it.Cur)}, map[string]string{"hari_produksi": fmtDay(it.Deadline)},
			"Jam kirim >= jam siap dan kapasitas hari kirim cukup; H-1 dilonggarkan untuk pesanan yang wajib H-1"); err != nil {
			return err
		}
		if err := s.notify(ctx, tx, it.OrderID, fmt.Sprintf(
			"Order %s digeser dari %s ke %s (hari kirim) agar kapasitas H-1 longgar untuk pesanan yang wajib H-1.",
			it.Kode, fmtDay(it.Cur), fmtDay(it.Deadline))); err != nil {
			return err
		}
	}

	// At risk
	if now.Hour() >= set.ProductionEndHour-set.WarnHours {
		risky, err := s.bakery.GetRiskOrders(ctx, tx, fmtDay(today))
		if err != nil {
			return err
		}
		for _, r := range risky {
			if err := s.bakery.UpdateOrderStatus(ctx, tx, r.ID, "at_risk"); err != nil {
				return err
			}
			if err := s.audit(ctx, tx, r.ID, "sistem", "at_risk", nil, nil, "Belum mulai diproduksi padahal jam produksi hari ini hampir selesai"); err != nil {
				return err
			}
			if err := s.notify(ctx, tx, r.ID, fmt.Sprintf("Order %s berisiko terlambat: belum mulai diproduksi.", r.Kode)); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}
