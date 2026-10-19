package bakery

import (
	"context"
	"strconv"
	"time"

	"bakery-scheduler/internal/entity/bakery"

	"github.com/jmoiron/sqlx"
)

// GetSettings membaca semua parameter (kunci -> nilai).
func (d Data) GetSettings(ctx context.Context, tx *sqlx.Tx) (map[string]float64, error) {
	rows, err := d.st(ctx, tx, getSettings).QueryxContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // wajib ditutup sebelum query berikutnya pada transaksi yang sama
	out := map[string]float64{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		f, _ := strconv.ParseFloat(v, 64)
		out[k] = f
	}
	return out, rows.Err()
}

// GetActiveWeights mengembalikan sql.ErrNoRows bila belum ada bobot aktif.
func (d Data) GetActiveWeights(ctx context.Context, tx *sqlx.Tx) (bakery.Weights, error) {
	var w bakery.Weights
	err := d.st(ctx, tx, getActiveWeights).QueryRowxContext(ctx).Scan(&w.J, &w.Q, &w.W)
	return w, err
}

func (d Data) CountOverReserved(ctx context.Context, tx *sqlx.Tx, today string, now time.Time, limit float64) (int, error) {
	var n int
	err := d.st(ctx, tx, countOverReserved).QueryRowxContext(ctx, today, now, limit).Scan(&n)
	return n, err
}

func (d Data) UpdateFutureCapacity(ctx context.Context, tx *sqlx.Tx, base, overtimeMax int, today string) error {
	_, err := d.st(ctx, tx, updFutureCapacity).ExecContext(ctx, base, overtimeMax, today)
	return err
}

func (d Data) UpsertSetting(ctx context.Context, tx *sqlx.Tx, key, value string) error {
	_, err := d.st(ctx, tx, upsertSetting).ExecContext(ctx, key, value)
	return err
}

func (d Data) DeactivateWeights(ctx context.Context, tx *sqlx.Tx) error {
	_, err := d.st(ctx, tx, deactivateWeights).ExecContext(ctx)
	return err
}

func (d Data) InsertWeights(ctx context.Context, tx *sqlx.Tx, matrixJSON string, w []float64, cr float64) error {
	_, err := d.st(ctx, tx, insWeights).ExecContext(ctx, matrixJSON, w[0], w[1], w[2], cr)
	return err
}
