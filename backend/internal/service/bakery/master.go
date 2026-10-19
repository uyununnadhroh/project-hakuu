package bakery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"bakery-scheduler/internal/entity/bakery"
)

// Bobot kapasitas bawaan per jenis produksi. Produk bertopping butuh waktu
// lebih lama (800 box/hari tanpa topping vs sekitar 650 box/hari bila
// didominasi topping, jadi 800/650 = 1,23).
var jenisBobot = map[string]float64{
	"tanpa_topping": 1.00,
	"bertopping":    1.23,
	"campuran":      1.10,
}

func normJenis(j string) (string, error) {
	j = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(j), " ", "_"))
	if j == "" {
		return "tanpa_topping", nil
	}
	if _, ok := jenisBobot[j]; !ok {
		return "", fmt.Errorf("%w: jenis harus tanpa_topping, bertopping, atau campuran", bakery.ErrValidation)
	}
	return j, nil
}

func checkBobot(b float64) error {
	if b < 0.1 || b > 5 {
		return fmt.Errorf("%w: bobot harus antara 0,1 dan 5", bakery.ErrValidation)
	}
	return nil
}

func (s Service) categoryExists(ctx context.Context, nama string) (bool, error) {
	n, err := s.bakery.CountCategory(ctx, nama)
	return n > 0, err
}

// ------------------------- PRODUK -------------------------

// ListProducts daftar produk; paket pilih varian ikut membawa daftar variannya.
func (s Service) ListProducts(ctx context.Context) ([]bakery.Product, error) {
	list, err := s.bakery.ListProducts(ctx)
	if err != nil {
		return nil, err
	}
	vars, err := s.bakery.ListProductVariants(ctx)
	if err != nil {
		return nil, err
	}
	idx := map[int64]int{}
	for i, p := range list {
		idx[p.ID] = i
	}
	for _, v := range vars {
		if i, ok := idx[v.ProductID]; ok {
			list[i].Variants = append(list[i].Variants, v)
		}
	}
	return list, nil
}

// AddProduct menambah produk lengkap dengan kategori, jenis, dan bobot.
func (s Service) AddProduct(ctx context.Context, in bakery.ProductInput) (int64, error) {
	in.Nama = strings.TrimSpace(in.Nama)
	if in.Nama == "" || in.Harga < 0 {
		return 0, fmt.Errorf("%w: nama wajib diisi dan harga tidak boleh negatif", bakery.ErrValidation)
	}
	if in.Kategori == "" {
		in.Kategori = "Menu Hemat"
	}
	ok, err := s.categoryExists(ctx, in.Kategori)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("%w: kategori %q belum ada", bakery.ErrValidation, in.Kategori)
	}
	jenis, err := normJenis(in.Jenis)
	if err != nil {
		return 0, err
	}
	bobot := jenisBobot[jenis]
	if in.Bobot != nil {
		bobot = *in.Bobot
	}
	if err := checkBobot(bobot); err != nil {
		return 0, err
	}
	return s.bakery.InsertProduct(ctx, in.Nama, in.Harga, in.Kategori, jenis, bobot)
}

// PatchProduct mengubah sebagian kolom produk (PATCH /api/products/{id}).
func (s Service) PatchProduct(ctx context.Context, id int64, in bakery.ReqPatchProduct) error {
	if in.Nama == nil && in.Harga == nil && in.Aktif == nil && in.Kategori == nil && in.Jenis == nil && in.Bobot == nil {
		return fmt.Errorf("%w: tidak ada yang diubah", bakery.ErrValidation)
	}
	if in.Nama != nil || in.Harga != nil || in.Aktif != nil {
		if err := s.UpdateProduct(ctx, id, in.Nama, in.Harga, in.Aktif); err != nil {
			return err
		}
	}
	return s.UpdateProductDetail(ctx, id, in.Kategori, in.Jenis, in.Bobot)
}

// UpdateProduct: ubah nama, harga, atau status aktif. Harga baru tidak
// mengubah order lama karena harga disalin ke order_items saat order dibuat.
func (s Service) UpdateProduct(ctx context.Context, id int64, nama *string, harga *float64, aktif *bool) error {
	if nama == nil && harga == nil && aktif == nil {
		return fmt.Errorf("%w: tidak ada yang diubah", bakery.ErrValidation)
	}
	if nama != nil && *nama == "" {
		return fmt.Errorf("%w: nama tidak boleh kosong", bakery.ErrValidation)
	}
	if harga != nil && *harga < 0 {
		return fmt.Errorf("%w: harga tidak boleh negatif", bakery.ErrValidation)
	}
	exists, err := s.bakery.CountProductByID(ctx, id)
	if err != nil {
		return err
	}
	if exists == 0 {
		return bakery.ErrNotFound
	}
	if nama != nil {
		if err := s.bakery.UpdateProductNama(ctx, id, *nama); err != nil {
			return err
		}
	}
	if harga != nil {
		if err := s.bakery.UpdateProductHarga(ctx, id, *harga); err != nil {
			return err
		}
	}
	if aktif != nil {
		if err := s.bakery.UpdateProductAktif(ctx, id, *aktif); err != nil {
			return err
		}
	}
	return nil
}

// UpdateProductDetail mengubah kategori, jenis, dan/atau bobot. Bila jenis
// diganti tanpa bobot, bobot mengikuti bawaan jenis itu.
func (s Service) UpdateProductDetail(ctx context.Context, id int64, kategori, jenis *string, bobot *float64) error {
	if kategori == nil && jenis == nil && bobot == nil {
		return nil
	}
	exists, err := s.bakery.CountProductByID(ctx, id)
	if err != nil {
		return err
	}
	if exists == 0 {
		return bakery.ErrNotFound
	}
	if kategori != nil {
		ok, err := s.categoryExists(ctx, *kategori)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%w: kategori %q belum ada", bakery.ErrValidation, *kategori)
		}
		if err := s.bakery.UpdateProductKategori(ctx, id, *kategori); err != nil {
			return err
		}
	}
	if jenis != nil {
		j, err := normJenis(*jenis)
		if err != nil {
			return err
		}
		b := jenisBobot[j]
		if bobot != nil {
			b = *bobot
		}
		if err := checkBobot(b); err != nil {
			return err
		}
		return s.bakery.UpdateProductJenisBobot(ctx, id, j, b)
	}
	if bobot != nil {
		if err := checkBobot(*bobot); err != nil {
			return err
		}
		return s.bakery.UpdateProductBobot(ctx, id, *bobot)
	}
	return nil
}

// DeleteProduct hanya boleh untuk produk yang belum pernah dipesan; selain itu
// riwayat order akan rusak, jadi produk cukup dinonaktifkan.
func (s Service) DeleteProduct(ctx context.Context, id int64) error {
	used, err := s.bakery.CountProductUsage(ctx, id)
	if err != nil {
		return err
	}
	if used > 0 {
		return fmt.Errorf("%w: produk sudah pernah dipesan, nonaktifkan saja", bakery.ErrState)
	}
	n, err := s.bakery.DeleteProduct(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return bakery.ErrNotFound
	}
	return nil
}

// ------------------------- KATEGORI -------------------------

func (s Service) ListCategories(ctx context.Context) ([]string, error) {
	return s.bakery.ListCategories(ctx)
}

func (s Service) AddCategory(ctx context.Context, nama string) error {
	nama = strings.TrimSpace(nama)
	if nama == "" || len([]rune(nama)) > 60 {
		return fmt.Errorf("%w: nama kategori wajib diisi (maksimal 60 karakter)", bakery.ErrValidation)
	}
	ok, err := s.categoryExists(ctx, nama)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("%w: kategori %q sudah ada", bakery.ErrValidation, nama)
	}
	return s.bakery.InsertCategory(ctx, nama)
}

// DeleteCategory ditolak bila masih ada produk di kategori itu.
func (s Service) DeleteCategory(ctx context.Context, nama string) error {
	used, err := s.bakery.CountProductByCategory(ctx, nama)
	if err != nil {
		return err
	}
	if used > 0 {
		return fmt.Errorf("%w: kategori masih dipakai %d produk", bakery.ErrState, used)
	}
	n, err := s.bakery.DeleteCategory(ctx, nama)
	if err != nil {
		return err
	}
	if n == 0 {
		return bakery.ErrNotFound
	}
	return nil
}

// ------------------------- PELANGGAN -------------------------

func (s Service) ListCustomers(ctx context.Context) ([]bakery.Customer, error) {
	return s.bakery.ListCustomers(ctx)
}

// FindOrCreateCustomer memakai pelanggan yang sudah ada (nama + telepon sama)
// agar pembeli yang sama tidak tercatat berkali-kali; alamat diperbarui bila diisi.
func (s Service) FindOrCreateCustomer(ctx context.Context, nama, telepon, alamat string) (int64, error) {
	if nama == "" {
		return 0, fmt.Errorf("%w: nama pelanggan wajib diisi", bakery.ErrValidation)
	}
	id, err := s.bakery.FindCustomer(ctx, nama, telepon)
	if err == nil {
		if alamat != "" {
			if err := s.bakery.UpdateCustomerAlamat(ctx, id, alamat); err != nil {
				return 0, err
			}
		}
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	return s.bakery.InsertCustomer(ctx, nama, telepon, alamat)
}
