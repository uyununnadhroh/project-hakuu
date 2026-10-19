package bakery

import "errors"

// Error standar. Handler memetakan error ini ke HTTP status:
// ErrValidation -> 400, ErrNotFound -> 404, ErrState/ErrExpired/ErrCapacity -> 409.
var (
	ErrValidation = errors.New("validasi gagal")
	ErrNotFound   = errors.New("data tidak ditemukan")
	ErrState      = errors.New("status tidak sesuai")
	ErrExpired    = errors.New("tawaran sudah kedaluwarsa")
	ErrCapacity   = errors.New("kapasitas tidak cukup")

	// ErrLoginFailed (401) pesan sama untuk email salah dan sandi salah supaya akun tidak bisa ditebak.
	ErrLoginFailed = errors.New("email atau kata sandi salah")
)
