package bakery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"bakery-scheduler/internal/entity/bakery"
)

func (s Service) ListOrders(ctx context.Context, status string, limit int) ([]bakery.OrderRow, error) {
	// laporan penjualan bulanan butuh daftar yang lebih panjang
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}
	list, err := s.bakery.ListOrders(ctx, status, limit)
	if err != nil {
		return nil, err
	}
	return applyLabels(list, fmtDay(s.now())), nil
}

func (s Service) GetOrder(ctx context.Context, id int64) (*bakery.OrderDetail, error) {
	list, err := s.bakery.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, bakery.ErrNotFound
	}
	list = applyLabels(list, fmtDay(s.now()))
	d := &bakery.OrderDetail{OrderRow: list[0]}

	ex, err := s.bakery.GetOrderExtra(ctx, id)
	if err != nil {
		return nil, err
	}
	d.Total, d.Phone, d.Address, d.Note = ex.Total, ex.Phone, ex.Address, ex.Note
	if ex.J != nil {
		d.ScoreParts = map[string]float64{"J": *ex.J, "Q": *ex.Q, "W": *ex.W}
	}
	if d.Items, err = s.bakery.GetOrderItems(ctx, id); err != nil {
		return nil, err
	}
	if d.Offers, err = s.bakery.GetOrderOffers(ctx, id); err != nil {
		return nil, err
	}
	if d.History, err = s.bakery.GetOrderHistory(ctx, id); err != nil {
		return nil, err
	}
	return d, nil
}

// Schedule menampilkan kapasitas dan order per hari produksi.
// from format YYYY-MM-DD (kosong = hari ini), days default 7 (maksimal 31).
func (s Service) Schedule(ctx context.Context, fromStr string, days int) ([]bakery.DayView, error) {
	now := s.now()
	from := now
	if fromStr != "" {
		t, err := time.ParseInLocation("2006-01-02", fromStr, s.loc)
		if err != nil {
			return nil, fmt.Errorf("%w: format from harus YYYY-MM-DD", bakery.ErrValidation)
		}
		from = t
	}
	set, err := s.loadSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	if days <= 0 || days > 31 {
		days = 7
	}
	out := []bakery.DayView{}
	for i := 0; i < days; i++ {
		d := AddDays(DayOf(from), i)
		ds := fmtDay(d)
		v := bakery.DayView{Date: ds, Capacity: set.CapacityPerDay}
		capacity, locked, err := s.bakery.GetDayCapacity(ctx, ds)
		if err == nil {
			v.Capacity, v.Locked = capacity, locked || DayLocked(d, now, set)
		} else if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if v.Used, v.Overtime, err = s.bakery.SumUsedOvertime(ctx, ds, now); err != nil {
			return nil, err
		}
		orders, err := s.bakery.GetScheduleOrders(ctx, ds)
		if err != nil {
			return nil, err
		}
		v.Orders = applyLabels(orders, fmtDay(now))
		out = append(out, v)
	}
	return out, nil
}
