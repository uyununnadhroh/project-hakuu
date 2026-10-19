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

func TestNormVariants(t *testing.T) {
	off := false
	v := func(nama string, opsi ...string) bakery.VariantInput {
		return bakery.VariantInput{Nama: nama, OpsiPilihan: opsi}
	}
	mati := bakery.VariantInput{Nama: "Abon", Aktif: &off}

	cases := []struct {
		name string
		maks int
		in   []bakery.VariantInput
		ok   bool
	}{
		{"produk biasa tanpa varian", 0, nil, true},
		{"produk biasa diberi varian", 0, []bakery.VariantInput{v("Coklat")}, false},
		{"isi paket negatif", -1, nil, false},
		{"isi paket terlalu besar", MaxIsiPaket + 1, nil, false},
		{"paket pas", 2, []bakery.VariantInput{v("Coklat"), v("Keju")}, true},
		{"varian aktif kurang", 3, []bakery.VariantInput{v("Coklat"), v("Keju"), mati}, false},
		{"nama kosong", 1, []bakery.VariantInput{v("  ")}, false},
		{"nama dobel beda huruf", 1, []bakery.VariantInput{v("Coklat"), v("coklat ")}, false},
		{"opsi valid", 1, []bakery.VariantInput{v("Lava", "Stroberi", "Nanas")}, true},
		{"opsi kosong", 1, []bakery.VariantInput{v("Lava", "Stroberi", " ")}, false},
		{"opsi dobel", 1, []bakery.VariantInput{v("Lava", "Nanas", "nanas")}, false},
	}
	for _, c := range cases {
		_, err := NormVariants(c.maks, c.in)
		if (err == nil) != c.ok {
			t.Errorf("%s: ok=%v, err=%v", c.name, c.ok, err)
		}
	}

	// dirapikan: nama/opsi di-trim, label bawaan "Pilihan", aktif bawaan true
	out, err := NormVariants(1, []bakery.VariantInput{{Nama: " Lava ", OpsiPilihan: []string{" Stroberi "}}, mati})
	if err != nil {
		t.Fatal(err)
	}
	if out[0].Nama != "Lava" || out[0].OpsiLabel != "Pilihan" || out[0].OpsiPilihan[0] != "Stroberi" || !out[0].Aktif || out[1].Aktif {
		t.Errorf("hasil tidak dirapikan: %+v", out)
	}
}
