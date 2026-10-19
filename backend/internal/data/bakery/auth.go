package bakery

import (
	"context"
	"time"

	"bakery-scheduler/internal/entity/bakery"
)

// GetUserByEmail mengembalikan sql.ErrNoRows bila email tidak terdaftar.
func (d Data) GetUserByEmail(ctx context.Context, email string) (bakery.User, error) {
	var u bakery.User
	err := (*d.stmt)[getUserByEmail].QueryRowxContext(ctx, email).Scan(&u.ID, &u.Nama, &u.Email, &u.PasswordHash, &u.Aktif)
	return u, err
}

// ListUsers akun aktif yang bisa masuk ke website (tanpa password).
func (d Data) ListUsers(ctx context.Context) ([]bakery.UserView, error) {
	rows, err := (*d.stmt)[listUsers].QueryxContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.UserView{}
	for rows.Next() {
		var u bakery.UserView
		var created time.Time
		if err := rows.Scan(&u.ID, &u.Nama, &u.Email, &created); err != nil {
			return nil, err
		}
		u.CreatedAt = created.Format("2006-01-02T15:04:05")
		out = append(out, u)
	}
	return out, rows.Err()
}
