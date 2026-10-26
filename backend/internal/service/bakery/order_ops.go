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
// kunci jadwal hari ini (frozen window) mulai jam kunci rencana (08.00).
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
	if DayLocked(now, now, set) {
		if err := s.bakery.LockDaysUntil(ctx, tx, fmtDay(DayOf(now))); err != nil {
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
//  1. Memajukan: order di dalam jendela 72 jam diurutkan dari skor tertinggi;
//     bila ada hari produksi yang lebih awal (dan masih diizinkan) dengan
//     kapasitas kosong (misalnya karena ada pembatalan), order dipindah ke sana.
//     Hari yang dikunci (frozen) dan order yang sedang diproduksi tidak disentuh.
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
		if InWindow(it.Deliver, now, set) {
			it.Score = ComputeScore(bakery.ScoreInput{Qty: it.Qty, DeliverAt: it.Deliver, OrderedAt: it.Order}, now, set).Total
			items = append(items, it)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Score > items[j].Score })

	for _, it := range items {
		for d := EarliestDay(it.Deliver, now, set); d.Before(it.Cur); d = AddDays(d, 1) {
			st, err := s.getDay(ctx, tx, d, set, now)
			if err != nil {
				return err
			}
			if st.Locked || st.Free() < it.Qty {
				continue
			}
			if err := s.bakery.UpdateReservationDay(ctx, tx, it.ResID, fmtDay(d)); err != nil {
				return err
			}
			if err := s.bakery.UpdateOrderProductionDay(ctx, tx, it.OrderID, fmtDay(d)); err != nil {
				return err
			}
			if err := s.audit(ctx, tx, it.OrderID, "sistem", "dimajukan",
				map[string]string{"hari_produksi": fmtDay(it.Cur)}, map[string]string{"hari_produksi": fmtDay(d)},
				fmt.Sprintf("Ada kapasitas kosong di hari lebih awal (skor %.3f)", it.Score)); err != nil {
				return err
			}
			if err := s.notify(ctx, tx, it.OrderID, fmt.Sprintf("Order %s dimajukan dari %s ke %s.", it.Kode, fmtDay(it.Cur), fmtDay(d))); err != nil {
				return err
			}
			break
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
