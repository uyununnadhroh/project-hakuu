package bakery

import (
	"context"
	"database/sql"
	"time"

	"bakery-scheduler/internal/entity/bakery"

	"github.com/jmoiron/sqlx"
)

func (d Data) InsertOffer(ctx context.Context, tx *sqlx.Tx, orderID int64, urutan int, jam time.Time, hariProduksi string) (int64, error) {
	r, err := d.st(ctx, tx, insOffer).ExecContext(ctx, orderID, urutan, jam, hariProduksi)
	if err != nil {
		return 0, err
	}
	return r.LastInsertId()
}

// GetOfferLock mengunci tawaran. sql.ErrNoRows bila tidak ada (ProdDay belum di-parse).
func (d Data) GetOfferLock(ctx context.Context, tx *sqlx.Tx, id int64) (bakery.OfferRow, error) {
	var o bakery.OfferRow
	err := d.st(ctx, tx, getOfferLock).QueryRowxContext(ctx, id).
		Scan(&o.ID, &o.OrderID, &o.Status, &o.DeliverAt, &o.ProdDayStr, &o.ResID, &o.Due, &o.Qty, &o.OrderStatus, &o.Kode, &o.OrderedAt)
	return o, err
}

func (d Data) CountActiveOffers(ctx context.Context, tx *sqlx.Tx, orderID int64, now time.Time) (int, error) {
	var n int
	err := d.st(ctx, tx, countActiveOffers).QueryRowxContext(ctx, orderID, now).Scan(&n)
	return n, err
}

func (d Data) UpdateOfferOffered(ctx context.Context, tx *sqlx.Tx, id int64, now, due time.Time, resID int64) error {
	_, err := d.st(ctx, tx, updOfferOffered).ExecContext(ctx, now, due, resID, id)
	return err
}

// UpdateOfferAnswered status: diterima / ditolak / kedaluwarsa. Reservasi dilepas dari tawaran.
func (d Data) UpdateOfferAnswered(ctx context.Context, tx *sqlx.Tx, id int64, status string, now time.Time) error {
	_, err := d.st(ctx, tx, updOfferAnswered).ExecContext(ctx, status, now, id)
	return err
}

func (d Data) UpdateOrderAfterNegotiation(ctx context.Context, tx *sqlx.Tx, orderID int64, deliverAt time.Time,
	deadline, prodDay, status string, sc bakery.Score, now time.Time) error {
	_, err := d.st(ctx, tx, updOrderAfterNego).ExecContext(ctx, deliverAt, deadline, prodDay, status,
		sc.J, sc.Q, sc.W, sc.Total, now, orderID)
	return err
}

func (d Data) ExpireRecommendedOffers(ctx context.Context, tx *sqlx.Tx, orderID int64) error {
	_, err := d.st(ctx, tx, expireRecommended).ExecContext(ctx, orderID)
	return err
}

func (d Data) ExpireOpenOffersByOrder(ctx context.Context, tx *sqlx.Tx, orderID int64) error {
	_, err := d.st(ctx, tx, expireOpenByOrder).ExecContext(ctx, orderID)
	return err
}

func (d Data) GetExpiredOffers(ctx context.Context, tx *sqlx.Tx, now time.Time) ([]bakery.ExpiredOffer, error) {
	rows, err := d.st(ctx, tx, getExpiredOffers).QueryxContext(ctx, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []bakery.ExpiredOffer
	for rows.Next() {
		var e bakery.ExpiredOffer
		var res sql.NullInt64
		if err := rows.Scan(&e.ID, &e.OrderID, &res); err != nil {
			return nil, err
		}
		e.ResID = res.Int64
		out = append(out, e)
	}
	return out, rows.Err()
}
