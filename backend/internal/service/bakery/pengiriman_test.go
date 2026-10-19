package bakery

import (
	"errors"
	"testing"
	"time"

	"bakery-scheduler/internal/entity/bakery"
)

func TestBatchKirim(t *testing.T) {
	at := func(h, m int) time.Time { return time.Date(2026, 10, 20, h, m, 0, 0, WIB) }
	cases := []struct {
		kirim time.Time
		want  string
	}{
		{at(7, 0), "pagi"},
		{at(10, 59), "pagi"},
		{at(11, 0), "malam"},
		{at(19, 0), "malam"},
	}
	for _, c := range cases {
		if got := BatchKirim(c.kirim); got != c.want {
			t.Errorf("kirim %s: batch %s, harusnya %s", c.kirim.Format("15:04"), got, c.want)
		}
	}
	if BatchDibuka("pagi", at(10, 59)) || !BatchDibuka("pagi", at(11, 0)) {
		t.Error("batch pagi harus dibuka tepat 11.00")
	}
	if BatchDibuka("malam", at(19, 59)) || !BatchDibuka("malam", at(20, 0)) {
		t.Error("batch malam harus dibuka tepat 20.00")
	}
}

func TestCheckDeliveryConfirm(t *testing.T) {
	at := func(d, h int) time.Time { return time.Date(2026, 10, d, h, 0, 0, 0, WIB) }
	cases := []struct {
		name       string
		status, to string
		kirim, now time.Time
		wantErr    error
	}{
		{"pagi sesudah 11.00 selesai", "terlambat", "selesai", at(20, 9), at(20, 11), nil},
		{"pagi sesudah 11.00 belum terkirim", "terlambat", "belum_terkirim", at(20, 9), at(20, 11), nil},
		{"pagi sebelum 11.00", "diproduksi", "selesai", at(20, 9), at(20, 10), bakery.ErrState},
		{"malam jam 15.00", "diproduksi", "selesai", at(20, 14), at(20, 15), bakery.ErrState},
		{"malam sesudah 20.00", "terlambat", "selesai", at(20, 14), at(20, 20), nil},
		{"kirim besok", "terjadwal", "selesai", at(21, 9), at(20, 21), bakery.ErrState},
		{"hari sebelumnya, belum terkirim -> selesai", "terlambat", "selesai", at(19, 14), at(20, 8), nil},
		{"sudah selesai tidak bisa diubah", "selesai", "belum_terkirim", at(20, 9), at(20, 12), bakery.ErrState},
		{"dibatalkan", "batal", "selesai", at(20, 9), at(20, 12), bakery.ErrState},
		{"belum dijadwalkan", "menunggu_negosiasi", "selesai", at(20, 9), at(20, 12), bakery.ErrState},
		{"status tidak dikenal", "terlambat", "hilang", at(20, 9), at(20, 12), bakery.ErrValidation},
	}
	for _, c := range cases {
		err := CheckDeliveryConfirm(c.status, c.to, c.kirim, c.now)
		if c.wantErr == nil && err != nil || c.wantErr != nil && !errors.Is(err, c.wantErr) {
			t.Errorf("%s: err=%v, harusnya %v", c.name, err, c.wantErr)
		}
	}
}

// Jam dari zona lain (mis. server UTC atau APP_TZ berbeda) tetap dihitung dalam WIB.
func TestPengirimanMemakaiWIB(t *testing.T) {
	utc := func(d, h, m int) time.Time { return time.Date(2026, 10, d, h, m, 0, 0, time.UTC) }
	if BatchKirim(utc(20, 3, 59)) != "pagi" || BatchKirim(utc(20, 4, 0)) != "malam" {
		t.Error("10.59 WIB harus batch pagi, 11.00 WIB batch malam")
	}
	if BatchDibuka("pagi", utc(20, 3, 59)) || !BatchDibuka("pagi", utc(20, 4, 0)) {
		t.Error("batch pagi harus dibuka tepat 11.00 WIB (04.00 UTC)")
	}
	if BatchDibuka("malam", utc(20, 12, 59)) || !BatchDibuka("malam", utc(20, 13, 0)) {
		t.Error("batch malam harus dibuka tepat 20.00 WIB (13.00 UTC)")
	}
	// 19 Okt 17.30 UTC = 20 Okt 00.30 WIB: pesanan kirim 20 Okt 09.00 WIB belum boleh dikonfirmasi,
	// pesanan kirim 19 Okt 15.00 WIB (kemarin menurut WIB) sudah boleh.
	now := utc(19, 17, 30)
	if err := CheckDeliveryConfirm("terlambat", "selesai", utc(20, 2, 0), now); !errors.Is(err, bakery.ErrState) {
		t.Errorf("kirim hari ini (WIB) sebelum batch dibuka harus ditolak, err=%v", err)
	}
	if err := CheckDeliveryConfirm("terlambat", "selesai", utc(19, 8, 0), now); err != nil {
		t.Errorf("kirim kemarin (WIB) harus boleh dikonfirmasi, err=%v", err)
	}
}
