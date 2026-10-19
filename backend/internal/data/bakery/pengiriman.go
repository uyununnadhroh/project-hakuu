package bakery

import (
	"context"
	"database/sql"
	"time"

	"bakery-scheduler/internal/entity/bakery"

	"github.com/jmoiron/sqlx"
)

// ------------------------- KONFIRMASI PENGIRIMAN -------------------------

// GetDeliveryOrders pesanan dengan waktu kirim di [from, to), ditambah pesanan di
// [pastFrom, from) yang belum selesai. Pesanan selesai tanpa catatan konfirmasi
// (mis. lewat "Tandai selesai") dianggap terkonfirmasi selesai.
func (d Data) GetDeliveryOrders(ctx context.Context, from, to, pastFrom time.Time) ([]bakery.DeliveryOrder, error) {
	rows, err := (*d.stmt)[getDeliveryOrders].QueryxContext(ctx, from, to, pastFrom, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.DeliveryOrder{}
	for rows.Next() {
		var o bakery.DeliveryOrder
		var delivered, waktu sql.NullTime
		var dcStatus, alasan, oleh sql.NullString
		if err := rows.Scan(&o.ID, &o.Kode, &o.Customer, &o.Products, &o.Kirim, &o.Status, &delivered,
			&dcStatus, &alasan, &waktu, &oleh); err != nil {
			return nil, err
		}
		switch {
		case dcStatus.Valid:
			o.Konfirmasi = &bakery.DeliveryConfirmation{Status: dcStatus.String, Alasan: alasan.String,
				At: waktu.Time, Oleh: oleh.String}
		case o.Status == "selesai":
			o.Konfirmasi = &bakery.DeliveryConfirmation{Status: "selesai", At: delivered.Time}
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// GetOrderDeliveryLock status & waktu kirim pesanan, dikunci untuk dikonfirmasi.
// Mengembalikan sql.ErrNoRows bila pesanan tidak ada.
func (d Data) GetOrderDeliveryLock(ctx context.Context, tx *sqlx.Tx, id int64) (string, time.Time, error) {
	var status string
	var kirim time.Time
	err := d.st(ctx, tx, getOrderDeliveryLock).QueryRowxContext(ctx, id).Scan(&status, &kirim)
	return status, kirim, err
}

func (d Data) InsertDeliveryConfirmation(ctx context.Context, tx *sqlx.Tx, orderID int64, status, alasan string, userID int64, at time.Time) error {
	_, err := d.st(ctx, tx, insDeliveryConfirmation).ExecContext(ctx, orderID, status, nullStr(alasan), nullID(userID), at)
	return err
}
