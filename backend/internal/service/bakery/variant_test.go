package bakery

import (
	"testing"

	"bakery-scheduler/internal/entity/bakery"
)

func TestCheckVariants(t *testing.T) {
	avail := []bakery.ProductVariant{
		{ID: 1, Nama: "Coklat"},
		{ID: 2, Nama: "Keju"},
		{ID: 3, Nama: "Lava", OpsiLabel: "Isian selai", OpsiPilihan: []string{"Stroberi", "Blueberry"}},
	}
	v := func(id int64, opsi string) bakery.ItemVariant { return bakery.ItemVariant{VariantID: id, Opsi: opsi} }

	cases := []struct {
		name   string
		maks   int
		chosen []bakery.ItemVariant
		ok     bool
	}{
		{"produk biasa tanpa varian", 0, nil, true},
		{"produk biasa membawa varian", 0, []bakery.ItemVariant{v(1, "")}, false},
		{"paket lengkap", 2, []bakery.ItemVariant{v(1, ""), v(2, "")}, true},
		{"paket kurang varian", 2, []bakery.ItemVariant{v(1, "")}, false},
		{"paket kelebihan varian", 2, []bakery.ItemVariant{v(1, ""), v(2, ""), v(3, "Stroberi")}, false},
		{"varian ganda", 2, []bakery.ItemVariant{v(1, ""), v(1, "")}, false},
		{"varian milik paket lain", 2, []bakery.ItemVariant{v(1, ""), v(9, "")}, false},
		{"lava dengan isian valid", 2, []bakery.ItemVariant{v(1, ""), v(3, "Blueberry")}, true},
		{"lava tanpa isian", 2, []bakery.ItemVariant{v(1, ""), v(3, "")}, false},
		{"lava isian tidak dikenal", 2, []bakery.ItemVariant{v(1, ""), v(3, "Durian")}, false},
		{"opsi di varian tanpa opsi", 2, []bakery.ItemVariant{v(1, "Stroberi"), v(2, "")}, false},
	}
	for _, c := range cases {
		err := CheckVariants(c.maks, avail, c.chosen)
		if (err == nil) != c.ok {
			t.Errorf("%s: ok=%v, err=%v", c.name, c.ok, err)
		}
	}
}
