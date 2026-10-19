package bakery

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"bakery-scheduler/internal/entity/bakery"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s Service) Login(ctx context.Context, req bakery.ReqLogin) (bakery.LoginResult, error) {
	u, err := s.bakery.GetUserByEmail(ctx, req.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return bakery.LoginResult{}, bakery.ErrLoginFailed
	}
	if err != nil {
		return bakery.LoginResult{}, err
	}
	if !u.Aktif || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		return bakery.LoginResult{}, bakery.ErrLoginFailed
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": strconv.FormatInt(u.ID, 10), "name": u.Nama, "email": u.Email,
		"exp": time.Now().Add(12 * time.Hour).Unix(),
	}).SignedString(s.secret)
	if err != nil {
		return bakery.LoginResult{}, err
	}
	return bakery.LoginResult{Token: tok, Nama: u.Nama, Email: u.Email}, nil
}

// ListUsers daftar email yang bisa masuk ke website (menu Pengaturan).
func (s Service) ListUsers(ctx context.Context) ([]bakery.UserView, error) {
	return s.bakery.ListUsers(ctx)
}
