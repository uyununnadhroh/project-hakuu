package bakery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"bakery-scheduler/internal/entity/bakery"
)

const fmtMenit = "2006-01-02 15:04"

// DeliveryConfirmations daftar pesanan yang perlu dikonfirmasi pengirimannya
// saat ini: batch pagi (mulai 11.00), batch malam (mulai 20.00), dan pesanan
// 7 hari terakhir yang belum dikonfirmasi atau belum terkirim. Semua jam dalam WIB.
func (s Service) DeliveryConfirmations(ctx context.Context) (*bakery.DeliveryView, error) {
	now := s.now().In(WIB)
	today := DayOf(now)
	list, err := s.bakery.GetDeliveryOrders(ctx, today, AddDays(today, 1), AddDays(today, -7))
	if err != nil {
		return nil, err
	}
	pagi := bakery.DeliveryBatch{Key: "pagi", Label: "Batch pagi", Keterangan: "Kirim hari ini sebelum 11.00", Orders: []bakery.DeliveryOrder{}}
	malam := bakery.DeliveryBatch{Key: "malam", Label: "Batch malam", Keterangan: "Kirim hari ini mulai 11.00", Orders: []bakery.DeliveryOrder{}}
	lama := bakery.DeliveryBatch{Key: "sebelumnya", Label: "Hari sebelumnya", Keterangan: "Belum dikonfirmasi atau belum terkirim", Orders: []bakery.DeliveryOrder{}}
	for _, o := range list {
		o.Kirim = o.Kirim.In(WIB)
		o.DeliverAt = o.Kirim.Format(fmtMenit)
		if k := o.Konfirmasi; k != nil && !k.At.IsZero() {
			k.Waktu = k.At.In(WIB).Format(fmtMenit)
		}
		switch {
		case o.Kirim.Before(today):
			lama.Orders = append(lama.Orders, o)
		case BatchKirim(o.Kirim) == "pagi":
			if BatchDibuka("pagi", now) {
				pagi.Orders = append(pagi.Orders, o)
			}
		default:
			if BatchDibuka("malam", now) {
				malam.Orders = append(malam.Orders, o)
			}
		}
	}
	view := &bakery.DeliveryView{Date: fmtDay(today), Batches: []bakery.DeliveryBatch{}}
	for _, b := range []bakery.DeliveryBatch{pagi, malam, lama} {
		if len(b.Orders) == 0 {
			continue
		}
		view.Batches = append(view.Batches, b)
		// notifikasi tetap aktif selama ada pesanan yang belum Selesai
		// (belum dikonfirmasi maupun "belum terkirim")
		for _, o := range b.Orders {
			if o.Konfirmasi == nil || o.Konfirmasi.Status != "selesai" {
				view.Pending++
			}
		}
	}
	return view, nil
}

// ConfirmDelivery menyimpan konfirmasi pengiriman satu pesanan. "selesai" juga
// menutup pesanan (status selesai + waktu diserahkan).
func (s Service) ConfirmDelivery(ctx context.Context, orderID int64, req bakery.ReqConfirmDelivery, actor bakery.Actor) error {
	to := strings.TrimSpace(req.Status)
	alasan := strings.TrimSpace(req.Alasan)
	if to == "selesai" {
		alasan = ""
	}
	if len([]rune(alasan)) > 255 {
		return fmt.Errorf("%w: alasan maksimal 255 karakter", bakery.ErrValidation)
	}
	now := s.now().In(WIB)
	tx, err := s.bakery.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	status, kirim, err := s.bakery.GetOrderDeliveryLock(ctx, tx, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return bakery.ErrNotFound
	}
	if err != nil {
		return err
	}
	if err := CheckDeliveryConfirm(status, to, kirim, now); err != nil {
		return err
	}
	after := status
	if to == "selesai" {
		after = "selesai"
		if err := s.bakery.UpdateOrderStatus(ctx, tx, orderID, after); err != nil {
			return err
		}
		if err := s.bakery.UpdateOrderDelivered(ctx, tx, orderID, now); err != nil {
			return err
		}
	}
	if err := s.bakery.InsertDeliveryConfirmation(ctx, tx, orderID, to, alasan, actor.UserID, now); err != nil {
		return err
	}
	if err := s.audit(ctx, tx, orderID, actor.Nama, "konfirmasi_"+to,
		map[string]string{"status": status}, map[string]string{"status": after}, alasan); err != nil {
		return err
	}
	return tx.Commit()
}
