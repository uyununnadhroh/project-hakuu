package bakery

import (
	"errors"
	"testing"

	"bakery-scheduler/internal/entity/bakery"
)

func TestHitungBobot(t *testing.T) {
	cases := []struct {
		jenis        string
		isi, topping int
		want         float64
	}{
		{"bertopping", 3, 3, 1.23}, // patokan
		{"bertopping", 2, 2, 0.82},
		{"bertopping", 4, 4, 1.64},
		{"bertopping", 6, 6, 2.46},
		{"tanpa_topping", 2, 0, 1.00},
		{"tanpa_topping", 6, 0, 1.00},
		{"campuran", 3, 1, 1.08},
		{"campuran", 3, 2, 1.15},
		{"campuran", 2, 1, 0.91},
		{"campuran", 4, 1, 1.16},
		{"campuran", 4, 2, 1.32},
		{"campuran", 6, 3, 1.73},
	}
	for _, c := range cases {
		if got := HitungBobot(c.jenis, c.isi, c.topping); got != c.want {
			t.Errorf("HitungBobot(%s, isi %d, topping %d) = %v, want %v", c.jenis, c.isi, c.topping, got, c.want)
		}
	}
}

func TestNormKomposisi(t *testing.T) {
	if isi, top, err := normKomposisi("bertopping", 2, 0); err != nil || isi != 2 || top != 2 {
		t.Errorf("bertopping: semua roti bertopping, got %d/%d %v", isi, top, err)
	}
	if _, top, err := normKomposisi("tanpa_topping", 4, 3); err != nil || top != 0 {
		t.Errorf("tanpa topping: roti bertopping harus 0, got %d %v", top, err)
	}
	for _, c := range [][2]int{{3, 0}, {3, 3}, {1, 1}} {
		if _, _, err := normKomposisi("campuran", c[0], c[1]); !errors.Is(err, bakery.ErrValidation) {
			t.Errorf("campuran isi %d topping %d harus ditolak, got %v", c[0], c[1], err)
		}
	}
	if _, _, err := normKomposisi("bertopping", 0, 0); !errors.Is(err, bakery.ErrValidation) {
		t.Errorf("isi 0 harus ditolak, got %v", err)
	}
}
