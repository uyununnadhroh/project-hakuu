package bakery

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// LockDay mengunci baris daily_capacity (SELECT ... FOR UPDATE). Mengembalikan
// sql.ErrNoRows bila baris hari itu belum ada.
func (d Data) LockDay(ctx context.Context, tx *sqlx.Tx, day string) (base, overtimeMax int, locked bool, err error) {
	err = d.st(ctx, tx, lockDay).QueryRowxContext(ctx, day).Scan(&base, &overtimeMax, &locked)
	return base, overtimeMax, locked, err
}

func (d Data) InsertDayIgnore(ctx context.Context, tx *sqlx.Tx, day string, base, overtimeMax int) error {
	_, err := d.st(ctx, tx, insDayIgnore).ExecContext(ctx, day, base, overtimeMax)
	return err
}

// SumUsed kapasitas terpakai (tahanan sementara yang kedaluwarsa tidak dihitung).
func (d Data) SumUsed(ctx context.Context, tx *sqlx.Tx, day string, now time.Time) (int, error) {
	var used int
	err := d.st(ctx, tx, sumUsed).QueryRowxContext(ctx, day, now).Scan(&used)
	return used, err
}

func (d Data) GetDayCapacity(ctx context.Context, day string) (capacity int, locked bool, err error) {
	err = (*d.stmt)[getDayCapacity].QueryRowxContext(ctx, day).Scan(&capacity, &locked)
	return capacity, locked, err
}

func (d Data) SumUsedOvertime(ctx context.Context, day string, now time.Time) (used, overtime int, err error) {
	err = (*d.stmt)[sumUsedOvertime].QueryRowxContext(ctx, day, now).Scan(&used, &overtime)
	return used, overtime, err
}

// InsertReservation tipe: 'tetap' atau 'tahanan_sementara'. expires nil = NULL.
func (d Data) InsertReservation(ctx context.Context, tx *sqlx.Tx, orderID int64, day string, qty, overtime int,
	tipe string, expires *time.Time) (int64, error) {
	var exp interface{}
	if expires != nil {
		exp = *expires
	}
	r, err := d.st(ctx, tx, insReservation).ExecContext(ctx, orderID, day, qty, overtime, tipe, exp)
	if err != nil {
		return 0, err
	}
	return r.LastInsertId()
}

func (d Data) UpdateReservationDay(ctx context.Context, tx *sqlx.Tx, resID int64, day string) error {
	_, err := d.st(ctx, tx, updReservationDay).ExecContext(ctx, day, resID)
	return err
}

func (d Data) DeleteReservation(ctx context.Context, tx *sqlx.Tx, resID int64) error {
	_, err := d.st(ctx, tx, delReservation).ExecContext(ctx, resID)
	return err
}

func (d Data) DeleteReservationsByOrder(ctx context.Context, tx *sqlx.Tx, orderID int64) error {
	_, err := d.st(ctx, tx, delReservationByOrder).ExecContext(ctx, orderID)
	return err
}

func (d Data) LockDaysUntil(ctx context.Context, tx *sqlx.Tx, day string) error {
	_, err := d.st(ctx, tx, lockDaysUntil).ExecContext(ctx, day)
	return err
}

// InsertAudit orderID 0 = tidak terkait order. before/after nil = NULL.
func (d Data) InsertAudit(ctx context.Context, tx *sqlx.Tx, orderID int64, aksi string, before, after *string, alasan, pelaku string) error {
	var b, a interface{}
	if before != nil {
		b = *before
	}
	if after != nil {
		a = *after
	}
	_, err := d.st(ctx, tx, insAudit).ExecContext(ctx, nullID(orderID), aksi, b, a, alasan, pelaku)
	return err
}

func (d Data) InsertNotification(ctx context.Context, tx *sqlx.Tx, orderID int64, pesan string) error {
	_, err := d.st(ctx, tx, insNotification).ExecContext(ctx, nullID(orderID), pesan)
	return err
}
