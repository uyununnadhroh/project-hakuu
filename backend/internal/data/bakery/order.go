package bakery

import (
	"context"
	"database/sql"
	"time"

	"bakery-scheduler/internal/entity/bakery"

	"github.com/jmoiron/sqlx"
)

// GetProductPrice mengembalikan sql.ErrNoRows bila produk tidak ada.
func (d Data) GetProductPrice(ctx context.Context, tx *sqlx.Tx, productID int64) (bakery.ProductPrice, error) {
	var p bakery.ProductPrice
	err := d.st(ctx, tx, getProductPrice).QueryRowxContext(ctx, productID).Scan(&p.Harga, &p.Aktif, &p.Bobot, &p.MaksVarian)
	return p, err
}

func (d Data) InsertOrder(ctx context.Context, tx *sqlx.Tx, o bakery.OrderInsert) (int64, error) {
	r, err := d.st(ctx, tx, insOrder).ExecContext(ctx, o.CustomerID, o.Jumlah, o.QtyBox, o.TotalHarga,
		o.WaktuPesan, o.WaktuKirim, o.WaktuKirim, o.HariDeadline, o.Catatan, nullID(o.CreatedBy))
	if err != nil {
		return 0, err
	}
	return r.LastInsertId()
}

func (d Data) UpdateOrderKode(ctx context.Context, tx *sqlx.Tx, id int64, kode string) error {
	_, err := d.st(ctx, tx, updOrderKode).ExecContext(ctx, kode, id)
	return err
}

func (d Data) InsertOrderItem(ctx context.Context, tx *sqlx.Tx, orderID, productID int64, qty int, price float64) (int64, error) {
	r, err := d.st(ctx, tx, insOrderItem).ExecContext(ctx, orderID, productID, qty, price, price*float64(qty))
	if err != nil {
		return 0, err
	}
	return r.LastInsertId()
}

func (d Data) InsertOrderItemVariant(ctx context.Context, tx *sqlx.Tx, orderItemID, variantID int64, opsi string) error {
	_, err := d.st(ctx, tx, insOrderItemVariant).ExecContext(ctx, orderItemID, variantID, nullStr(opsi))
	return err
}

func (d Data) UpdateOrderNegotiation(ctx context.Context, tx *sqlx.Tx, id int64, sc bakery.Score, now time.Time) error {
	_, err := d.st(ctx, tx, updOrderNegotiation).ExecContext(ctx, sc.J, sc.Q, sc.W, sc.Total, now, id)
	return err
}

func (d Data) UpdateOrderProductionDay(ctx context.Context, tx *sqlx.Tx, id int64, day string) error {
	_, err := d.st(ctx, tx, updOrderProductionDay).ExecContext(ctx, day, id)
	return err
}

func (d Data) UpdateOrderAccepted(ctx context.Context, tx *sqlx.Tx, id int64, status, day string, overtime int,
	sc bakery.Score, now time.Time) error {
	_, err := d.st(ctx, tx, updOrderAccepted).ExecContext(ctx, status, day, overtime, sc.J, sc.Q, sc.W, sc.Total, now, id)
	return err
}

// GetCandidates order lain di hari itu yang masih boleh digeser (Deadline belum di-parse).
func (d Data) GetCandidates(ctx context.Context, tx *sqlx.Tx, day string) ([]bakery.Candidate, error) {
	rows, err := d.st(ctx, tx, getCandidates).QueryxContext(ctx, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []bakery.Candidate
	for rows.Next() {
		var c bakery.Candidate
		if err := rows.Scan(&c.ResID, &c.OrderID, &c.Kode, &c.Qty, &c.DeliverAt, &c.Ordered, &c.DeadlineStr); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetOrderStatusLock mengunci baris order. sql.ErrNoRows bila tidak ada.
func (d Data) GetOrderStatusLock(ctx context.Context, tx *sqlx.Tx, id int64) (string, error) {
	var status string
	err := d.st(ctx, tx, getOrderStatusLock).QueryRowxContext(ctx, id).Scan(&status)
	return status, err
}

func (d Data) GetOrderProductionDay(ctx context.Context, tx *sqlx.Tx, id int64) (sql.NullString, error) {
	var pd sql.NullString
	err := d.st(ctx, tx, getOrderProdDay).QueryRowxContext(ctx, id).Scan(&pd)
	return pd, err
}

func (d Data) UpdateOrderStatus(ctx context.Context, tx *sqlx.Tx, id int64, status string) error {
	_, err := d.st(ctx, tx, updOrderStatus).ExecContext(ctx, status, id)
	return err
}

func (d Data) UpdateOrderDelivered(ctx context.Context, tx *sqlx.Tx, id int64, at time.Time) error {
	_, err := d.st(ctx, tx, updOrderDelivered).ExecContext(ctx, at, id)
	return err
}

func (d Data) GetNewOrdersInWindow(ctx context.Context, tx *sqlx.Tx, until time.Time) ([]int64, error) {
	rows, err := d.st(ctx, tx, getNewOrdersInWindow).QueryxContext(ctx, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (d Data) MarkLateOrders(ctx context.Context, tx *sqlx.Tx, now time.Time) error {
	_, err := d.st(ctx, tx, updLateOrders).ExecContext(ctx, now)
	return err
}

func (d Data) GetOrdersForRescore(ctx context.Context, tx *sqlx.Tx) ([]bakery.ScoreRow, error) {
	rows, err := d.st(ctx, tx, getOrdersForRescore).QueryxContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []bakery.ScoreRow
	for rows.Next() {
		var r bakery.ScoreRow
		if err := rows.Scan(&r.ID, &r.Input.Qty, &r.Input.DeliverAt, &r.Input.OrderedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (d Data) UpdateOrderScore(ctx context.Context, tx *sqlx.Tx, id int64, sc bakery.Score, now time.Time) error {
	_, err := d.st(ctx, tx, updOrderScore).ExecContext(ctx, sc.J, sc.Q, sc.W, sc.Total, now, id)
	return err
}

// GetRebalanceItems order terjadwal setelah hari ini (Deadline/Cur belum di-parse).
func (d Data) GetRebalanceItems(ctx context.Context, tx *sqlx.Tx, today string) ([]bakery.RebalanceItem, error) {
	rows, err := d.st(ctx, tx, getRebalanceItems).QueryxContext(ctx, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []bakery.RebalanceItem
	for rows.Next() {
		var it bakery.RebalanceItem
		if err := rows.Scan(&it.ResID, &it.OrderID, &it.Kode, &it.Qty, &it.Deliver, &it.Order, &it.DeadlineStr, &it.CurStr); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (d Data) GetRiskOrders(ctx context.Context, tx *sqlx.Tx, today string) ([]bakery.RiskOrder, error) {
	rows, err := d.st(ctx, tx, getRiskOrders).QueryxContext(ctx, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []bakery.RiskOrder
	for rows.Next() {
		var r bakery.RiskOrder
		if err := rows.Scan(&r.ID, &r.Kode); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// scanOrders membaca hasil query ber-kolom orderCols. StatusLabel, NeedsNego,
// dan CanStart diisi di service (butuh tanggal hari ini).
func scanOrders(rows *sqlx.Rows) ([]bakery.OrderRow, error) {
	defer rows.Close()
	out := []bakery.OrderRow{}
	for rows.Next() {
		var o bakery.OrderRow
		var deliver, ordered time.Time
		var cust, pd, prods sql.NullString
		var sc sql.NullFloat64
		if err := rows.Scan(&o.ID, &o.Kode, &cust, &o.Qty, &o.QtyBox, &prods, &ordered, &deliver, &pd, &o.Status, &sc, &o.Overtime,
			&o.CreatedByNama, &o.CreatedByEmail); err != nil {
			return nil, err
		}
		o.DeliverAt = deliver.Format("2006-01-02 15:04")
		o.OrderedAt = ordered.Format("2006-01-02 15:04")
		o.Products = prods.String
		if cust.Valid {
			v := cust.String
			o.Customer = &v
		}
		if pd.Valid {
			v := pd.String
			o.ProductionDay = &v
		}
		if sc.Valid {
			v := sc.Float64
			o.Score = &v
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (d Data) ListOrders(ctx context.Context, status string, limit int) ([]bakery.OrderRow, error) {
	rows, err := (*d.stmt)[listOrders].QueryxContext(ctx, status, status, limit)
	if err != nil {
		return nil, err
	}
	return scanOrders(rows)
}

func (d Data) GetOrderByID(ctx context.Context, id int64) ([]bakery.OrderRow, error) {
	rows, err := (*d.stmt)[getOrderByID].QueryxContext(ctx, id)
	if err != nil {
		return nil, err
	}
	return scanOrders(rows)
}

func (d Data) GetScheduleOrders(ctx context.Context, day string) ([]bakery.OrderRow, error) {
	rows, err := (*d.stmt)[getScheduleOrders].QueryxContext(ctx, day)
	if err != nil {
		return nil, err
	}
	return scanOrders(rows)
}

func (d Data) GetOrderExtra(ctx context.Context, id int64) (bakery.OrderExtra, error) {
	var e bakery.OrderExtra
	var j, q, w sql.NullFloat64
	var phone, addr, note sql.NullString
	err := (*d.stmt)[getOrderExtra].QueryRowxContext(ctx, id).Scan(&e.Total, &j, &q, &w, &phone, &addr, &note)
	if err != nil {
		return e, err
	}
	e.Phone, e.Address, e.Note = phone.String, addr.String, note.String
	if j.Valid {
		jv, qv, wv := j.Float64, q.Float64, w.Float64
		e.J, e.Q, e.W = &jv, &qv, &wv
	}
	return e, nil
}

func (d Data) GetOrderItems(ctx context.Context, id int64) ([]bakery.OrderItemView, error) {
	rows, err := (*d.stmt)[getOrderItems].QueryxContext(ctx, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.OrderItemView{}
	for rows.Next() {
		var it bakery.OrderItemView
		if err := rows.Scan(&it.ID, &it.Produk, &it.Qty, &it.HargaSatuan, &it.Subtotal); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// varian paket (bila ada) ditempel ke baris item-nya
	vrows, err := (*d.stmt)[getOrderItemVariants].QueryxContext(ctx, id)
	if err != nil {
		return nil, err
	}
	defer vrows.Close()
	idx := map[int64]int{}
	for i, it := range out {
		idx[it.ID] = i
	}
	for vrows.Next() {
		var itemID int64
		var v bakery.OrderItemVariantView
		if err := vrows.Scan(&itemID, &v.Nama, &v.Opsi); err != nil {
			return nil, err
		}
		if i, ok := idx[itemID]; ok {
			out[i].Variants = append(out[i].Variants, v)
		}
	}
	return out, vrows.Err()
}

func (d Data) GetOrderOffers(ctx context.Context, id int64) ([]bakery.OrderOfferView, error) {
	rows, err := (*d.stmt)[getOrderOffers].QueryxContext(ctx, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.OrderOfferView{}
	for rows.Next() {
		var o bakery.OrderOfferView
		var t time.Time
		var due sql.NullTime
		if err := rows.Scan(&o.ID, &o.Seq, &t, &o.ProductionDay, &o.Status, &due); err != nil {
			return nil, err
		}
		o.DeliverAt = t.Format("2006-01-02 15:04")
		if due.Valid {
			o.RespondBefore = due.Time.Format(time.RFC3339)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (d Data) GetOrderHistory(ctx context.Context, id int64) ([]bakery.OrderHistoryView, error) {
	rows, err := (*d.stmt)[getOrderHistory].QueryxContext(ctx, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.OrderHistoryView{}
	for rows.Next() {
		var h bakery.OrderHistoryView
		var t time.Time
		if err := rows.Scan(&t, &h.Aksi, &h.Alasan, &h.Pelaku); err != nil {
			return nil, err
		}
		h.Waktu = t.Format("2006-01-02 15:04")
		out = append(out, h)
	}
	return out, rows.Err()
}
