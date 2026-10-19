package bakery

import (
	"context"
	"database/sql"
	"time"

	"bakery-scheduler/internal/entity/bakery"
)

func (d Data) GetSummary(ctx context.Context, from, to time.Time) (bakery.Summary, error) {
	var m bakery.Summary
	err := (*d.stmt)[getSummary].QueryRowxContext(ctx, from, to).Scan(&m.Orders, &m.Pcs, &m.Revenue)
	return m, err
}

func (d Data) GetDailyRevenue(ctx context.Context, from, to time.Time) ([]bakery.DayRevenue, error) {
	rows, err := (*d.stmt)[getDailyRevenue].QueryxContext(ctx, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.DayRevenue{}
	for rows.Next() {
		var r bakery.DayRevenue
		if err := rows.Scan(&r.Date, &r.Orders, &r.Pcs, &r.Revenue); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (d Data) GetTopProducts(ctx context.Context, from, to time.Time) ([]bakery.ProductRow, error) {
	rows, err := (*d.stmt)[getTopProducts].QueryxContext(ctx, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.ProductRow{}
	for rows.Next() {
		var p bakery.ProductRow
		if err := rows.Scan(&p.Name, &p.Pcs, &p.Revenue); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (d Data) GetTopCustomers(ctx context.Context, from, to time.Time) ([]bakery.CustomerRow, error) {
	rows, err := (*d.stmt)[getTopCustomers].QueryxContext(ctx, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.CustomerRow{}
	for rows.Next() {
		var c bakery.CustomerRow
		if err := rows.Scan(&c.Name, &c.Orders, &c.Spend); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (d Data) GetStatusCounts(ctx context.Context, from, to time.Time) (map[string]int, error) {
	rows, err := (*d.stmt)[getStatusCounts].QueryxContext(ctx, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var st string
		var n int
		if err := rows.Scan(&st, &n); err != nil {
			return nil, err
		}
		out[st] = n
	}
	return out, rows.Err()
}

func (d Data) GetDeliveredStats(ctx context.Context, from, to time.Time) (delivered, onTime int, err error) {
	err = (*d.stmt)[getDeliveredStats].QueryRowxContext(ctx, from, to).Scan(&delivered, &onTime)
	return delivered, onTime, err
}

func (d Data) GetUtilization(ctx context.Context, fromDay, toDay string) (sql.NullFloat64, error) {
	var u sql.NullFloat64
	err := (*d.stmt)[getUtilization].QueryRowxContext(ctx, fromDay, toDay).Scan(&u)
	return u, err
}

func (d Data) GetOvertimeStats(ctx context.Context, fromDay, toDay string) (days, pcs int, err error) {
	err = (*d.stmt)[getOvertimeStats].QueryRowxContext(ctx, fromDay, toDay).Scan(&days, &pcs)
	return days, pcs, err
}

func (d Data) GetAuditCounts(ctx context.Context, from, to time.Time) ([]bakery.AuditCount, error) {
	rows, err := (*d.stmt)[getAuditCounts].QueryxContext(ctx, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []bakery.AuditCount
	for rows.Next() {
		var a bakery.AuditCount
		if err := rows.Scan(&a.Aksi, &a.Total); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// GetMonthlyReport snapshot laporan yang sudah ditutup. sql.ErrNoRows bila belum ditutup.
func (d Data) GetMonthlyReport(ctx context.Context, bulan string) (string, time.Time, error) {
	var raw string
	var at time.Time
	err := (*d.stmt)[getMonthlyReport].QueryRowxContext(ctx, bulan).Scan(&raw, &at)
	return raw, at, err
}

func (d Data) CountMonthlyReport(ctx context.Context, bulan string) (int, error) {
	var n int
	err := (*d.stmt)[countMonthlyReport].QueryRowxContext(ctx, bulan).Scan(&n)
	return n, err
}

func (d Data) InsertMonthlyReport(ctx context.Context, bulan, data string) error {
	_, err := (*d.stmt)[insMonthlyReport].ExecContext(ctx, bulan, data)
	return err
}
