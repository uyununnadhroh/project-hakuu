package bakery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"bakery-scheduler/internal/entity/bakery"

	"github.com/jmoiron/sqlx"
)

func (s Service) loadOffer(ctx context.Context, tx *sqlx.Tx, id int64) (*bakery.OfferRow, error) {
	o, err := s.bakery.GetOfferLock(ctx, tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, bakery.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if o.ProdDay, err = s.parseDay(o.ProdDayStr); err != nil {
		return nil, err
	}
	return &o, nil
}

// OfferToBuyer dipanggil saat penjual memilih satu jam untuk ditawarkan ke
// pembeli. Kapasitas ditahan sementara selama HoldMinutes.
func (s Service) OfferToBuyer(ctx context.Context, offerID int64) (time.Time, error) {
	return retryDeadlock(func() (time.Time, error) { return s.offerOnce(ctx, offerID) })
}

func (s Service) offerOnce(ctx context.Context, offerID int64) (time.Time, error) {
	now := s.now()
	tx, err := s.bakery.BeginTx(ctx)
	if err != nil {
		return time.Time{}, err
	}
	defer tx.Rollback()
	set, err := s.loadSettings(ctx, tx)
	if err != nil {
		return time.Time{}, err
	}
	o, err := s.loadOffer(ctx, tx, offerID)
	if err != nil {
		return time.Time{}, err
	}
	if o.OrderStatus != "menunggu_negosiasi" || o.Status != "direkomendasikan" {
		return time.Time{}, fmt.Errorf("%w: tawaran tidak bisa diajukan", bakery.ErrState)
	}
	active, err := s.bakery.CountActiveOffers(ctx, tx, o.OrderID, now)
	if err != nil {
		return time.Time{}, err
	}
	if active > 0 {
		return time.Time{}, fmt.Errorf("%w: masih ada tawaran aktif untuk order ini", bakery.ErrState)
	}
	st, err := s.getDay(ctx, tx, o.ProdDay, set, now)
	if err != nil {
		return time.Time{}, err
	}
	if st.Free() < o.Qty {
		return time.Time{}, bakery.ErrCapacity
	}
	due := now.Add(time.Duration(set.HoldMinutes) * time.Minute)
	resID, err := s.bakery.InsertReservation(ctx, tx, o.OrderID, fmtDay(o.ProdDay), o.Qty, 0, "tahanan_sementara", &due)
	if err != nil {
		return time.Time{}, err
	}
	if err := s.bakery.UpdateOfferOffered(ctx, tx, o.ID, now, due, resID); err != nil {
		return time.Time{}, err
	}
	if err := s.audit(ctx, tx, o.OrderID, "penjual", "negosiasi_ditawarkan", nil,
		map[string]interface{}{"jam": o.DeliverAt.Format("2006-01-02 15:04"), "batas_respons": due.Format(time.RFC3339)},
		"Kapasitas ditahan sementara"); err != nil {
		return time.Time{}, err
	}
	return due, tx.Commit()
}

// RespondOffer mencatat jawaban pembeli. Bila setuju, kapasitas DICEK ULANG di
// dalam transaksi sebelum jam kirim order diubah.
func (s Service) RespondOffer(ctx context.Context, offerID int64, accept bool) error {
	_, err := retryDeadlock(func() (struct{}, error) { return struct{}{}, s.respondOnce(ctx, offerID, accept) })
	return err
}

func (s Service) respondOnce(ctx context.Context, offerID int64, accept bool) error {
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
	o, err := s.loadOffer(ctx, tx, offerID)
	if err != nil {
		return err
	}
	if o.Status != "ditawarkan" || !o.ResID.Valid {
		return fmt.Errorf("%w: tawaran belum diajukan atau sudah dijawab", bakery.ErrState)
	}
	release := func(status string) error {
		if err := s.bakery.DeleteReservation(ctx, tx, o.ResID.Int64); err != nil {
			return err
		}
		return s.bakery.UpdateOfferAnswered(ctx, tx, o.ID, status, now)
	}

	if o.Due.Valid && !o.Due.Time.After(now) {
		if err := release("kedaluwarsa"); err != nil {
			return err
		}
		if err := s.audit(ctx, tx, o.OrderID, "sistem", "negosiasi_kedaluwarsa", nil, nil, "Batas waktu respons lewat"); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		return bakery.ErrExpired
	}
	if !accept {
		if err := release("ditolak"); err != nil {
			return err
		}
		if err := s.audit(ctx, tx, o.OrderID, "penjual", "negosiasi_ditolak", nil, nil, "Pembeli menolak jam alternatif"); err != nil {
			return err
		}
		return tx.Commit()
	}

	if _, err := s.getDay(ctx, tx, o.ProdDay, set, now); err != nil { // kunci hari
		return err
	}
	if err := release("diterima"); err != nil { // lepas tahanan, lalu cek ulang
		return err
	}
	st, err := s.getDay(ctx, tx, o.ProdDay, set, now)
	if err != nil {
		return err
	}
	if st.Free() < o.Qty {
		return bakery.ErrCapacity // rollback: tahanan tetap utuh
	}
	if _, err := s.bakery.InsertReservation(ctx, tx, o.OrderID, fmtDay(o.ProdDay), o.Qty, 0, "tetap", nil); err != nil {
		return err
	}
	sc := ComputeScore(bakery.ScoreInput{Qty: o.Qty, DeliverAt: o.DeliverAt, OrderedAt: o.OrderedAt}, now, set)
	status := "terjadwal"
	if !InWindow(o.DeliverAt, now, set) {
		status = "baru"
	}
	if err := s.bakery.UpdateOrderAfterNegotiation(ctx, tx, o.OrderID, o.DeliverAt,
		fmtDay(DeadlineDay(o.DeliverAt, set)), fmtDay(o.ProdDay), status, sc, now); err != nil {
		return err
	}
	if err := s.bakery.ExpireRecommendedOffers(ctx, tx, o.OrderID); err != nil {
		return err
	}
	if err := s.audit(ctx, tx, o.OrderID, "penjual", "negosiasi_diterima", nil,
		map[string]string{"waktu_kirim": o.DeliverAt.Format("2006-01-02 15:04"), "hari_produksi": fmtDay(o.ProdDay)},
		"Pembeli setuju jam alternatif; kapasitas dicek ulang"); err != nil {
		return err
	}
	return tx.Commit()
}
