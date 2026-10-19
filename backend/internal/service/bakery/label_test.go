package bakery

import "testing"

func TestStatusLabel(t *testing.T) {
	d := func(s string) *string { return &s }
	cases := []struct {
		status string
		prod   *string
		want   string
	}{
		{"baru", nil, "Baru"},
		{"menunggu_negosiasi", nil, "Baru"},
		{"terjadwal", d("2026-10-14"), "Terjadwal 14 Okt"},
		{"terjadwal", d("2026-10-12"), "Diproduksi hari ini"}, // hari produksi = hari ini
		{"diproduksi", d("2026-10-12"), "Diproduksi hari ini"},
		{"at_risk", d("2026-10-12"), "Berisiko"},
		{"selesai", nil, "Selesai"},
	}
	for _, c := range cases {
		if got := statusLabel(c.status, c.prod, "2026-10-12"); got != c.want {
			t.Errorf("%s/%v = %q, want %q", c.status, c.prod, got, c.want)
		}
	}
}
