package bakery

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"bakery-scheduler/internal/entity/bakery"

	"github.com/jmoiron/sqlx"
)

// ------------------------- PRODUK -------------------------

func (d Data) ListProducts(ctx context.Context) ([]bakery.Product, error) {
	rows, err := (*d.stmt)[listProducts].QueryxContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.Product{}
	for rows.Next() {
		var p bakery.Product
		if err := rows.Scan(&p.ID, &p.Nama, &p.Harga, &p.Aktif, &p.Kategori, &p.Jenis, &p.Bobot, &p.MaksVarian); err != nil {
			return nil, err
		}
		p.Variants = []bakery.ProductVariant{}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListProductVariants semua varian aktif (untuk ditempel ke daftar produk).
func (d Data) ListProductVariants(ctx context.Context) ([]bakery.ProductVariant, error) {
	rows, err := (*d.stmt)[listProductVariants].QueryxContext(ctx)
	if err != nil {
		return nil, err
	}
	return scanVariants(rows)
}

// GetVariantsByProduct varian aktif milik satu produk (dipakai validasi order).
func (d Data) GetVariantsByProduct(ctx context.Context, tx *sqlx.Tx, productID int64) ([]bakery.ProductVariant, error) {
	rows, err := d.st(ctx, tx, getVariantsByProduct).QueryxContext(ctx, productID)
	if err != nil {
		return nil, err
	}
	return scanVariants(rows)
}

func scanVariants(rows *sqlx.Rows) ([]bakery.ProductVariant, error) {
	defer rows.Close()
	out := []bakery.ProductVariant{}
	for rows.Next() {
		var v bakery.ProductVariant
		var opsi []byte
		if err := rows.Scan(&v.ID, &v.ProductID, &v.Nama, &v.OpsiLabel, &opsi, &v.Aktif); err != nil {
			return nil, err
		}
		if len(opsi) > 0 {
			if err := json.Unmarshal(opsi, &v.OpsiPilihan); err != nil {
				return nil, err
			}
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// GetAllVariantsLock semua varian satu produk (termasuk nonaktif), dikunci untuk diubah.
func (d Data) GetAllVariantsLock(ctx context.Context, tx *sqlx.Tx, productID int64) ([]bakery.ProductVariant, error) {
	rows, err := d.st(ctx, tx, getAllVariantsLock).QueryxContext(ctx, productID)
	if err != nil {
		return nil, err
	}
	return scanVariants(rows)
}

func (d Data) CountVariantUsage(ctx context.Context, tx *sqlx.Tx, variantID int64) (int, error) {
	var n int
	err := d.st(ctx, tx, countVariantUsage).QueryRowxContext(ctx, variantID).Scan(&n)
	return n, err
}

// opsiJSON: pilihan opsi disimpan sebagai JSON array, kosong = NULL.
func opsiJSON(opsi []string) (interface{}, error) {
	if len(opsi) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(opsi)
	return string(b), err
}

func (d Data) InsertProductVariant(ctx context.Context, tx *sqlx.Tx, productID int64, v bakery.ProductVariant, urutan int) error {
	opsi, err := opsiJSON(v.OpsiPilihan)
	if err != nil {
		return err
	}
	_, err = d.st(ctx, tx, insProductVariant).ExecContext(ctx, productID, v.Nama, nullStr(v.OpsiLabel), opsi, v.Aktif, urutan)
	return err
}

func (d Data) UpdateProductVariant(ctx context.Context, tx *sqlx.Tx, productID int64, v bakery.ProductVariant, urutan int) error {
	opsi, err := opsiJSON(v.OpsiPilihan)
	if err != nil {
		return err
	}
	_, err = d.st(ctx, tx, updProductVariant).ExecContext(ctx, v.Nama, nullStr(v.OpsiLabel), opsi, v.Aktif, urutan, v.ID, productID)
	return err
}

func (d Data) DeleteProductVariant(ctx context.Context, tx *sqlx.Tx, id int64) error {
	_, err := d.st(ctx, tx, delProductVariant).ExecContext(ctx, id)
	return err
}

func (d Data) UpdateProductMaksVarian(ctx context.Context, tx *sqlx.Tx, id int64, maks int) error {
	_, err := d.st(ctx, tx, updProductMaksVarian).ExecContext(ctx, maks, id)
	return err
}

func (d Data) CountProductByID(ctx context.Context, id int64) (int, error) {
	var n int
	err := (*d.stmt)[countProductByID].QueryRowxContext(ctx, id).Scan(&n)
	return n, err
}

func (d Data) InsertProduct(ctx context.Context, tx *sqlx.Tx, nama string, harga float64, kategori, jenis string, bobot float64, maksVarian int) (int64, error) {
	r, err := d.st(ctx, tx, insProduct).ExecContext(ctx, nama, harga, kategori, jenis, bobot, maksVarian)
	if err != nil {
		return 0, err
	}
	return r.LastInsertId()
}

func (d Data) UpdateProductNama(ctx context.Context, id int64, nama string) error {
	_, err := (*d.stmt)[updProductNama].ExecContext(ctx, nama, id)
	return err
}

func (d Data) UpdateProductHarga(ctx context.Context, id int64, harga float64) error {
	_, err := (*d.stmt)[updProductHarga].ExecContext(ctx, harga, id)
	return err
}

func (d Data) UpdateProductAktif(ctx context.Context, id int64, aktif bool) error {
	_, err := (*d.stmt)[updProductAktif].ExecContext(ctx, aktif, id)
	return err
}

func (d Data) UpdateProductKategori(ctx context.Context, id int64, kategori string) error {
	_, err := (*d.stmt)[updProductKategori].ExecContext(ctx, kategori, id)
	return err
}

func (d Data) UpdateProductJenisBobot(ctx context.Context, id int64, jenis string, bobot float64) error {
	_, err := (*d.stmt)[updProductJenisBobot].ExecContext(ctx, jenis, bobot, id)
	return err
}

func (d Data) UpdateProductBobot(ctx context.Context, id int64, bobot float64) error {
	_, err := (*d.stmt)[updProductBobot].ExecContext(ctx, bobot, id)
	return err
}

func (d Data) CountProductUsage(ctx context.Context, id int64) (int, error) {
	var n int
	err := (*d.stmt)[countProductUsage].QueryRowxContext(ctx, id).Scan(&n)
	return n, err
}

// DeleteProduct mengembalikan jumlah baris yang terhapus.
func (d Data) DeleteProduct(ctx context.Context, id int64) (int64, error) {
	r, err := (*d.stmt)[delProduct].ExecContext(ctx, id)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

// ------------------------- KATEGORI -------------------------

func (d Data) ListCategories(ctx context.Context) ([]string, error) {
	rows, err := (*d.stmt)[listCategories].QueryxContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (d Data) CountCategory(ctx context.Context, nama string) (int, error) {
	var n int
	err := (*d.stmt)[countCategory].QueryRowxContext(ctx, nama).Scan(&n)
	return n, err
}

func (d Data) InsertCategory(ctx context.Context, nama string) error {
	_, err := (*d.stmt)[insCategory].ExecContext(ctx, nama)
	return err
}

func (d Data) CountProductByCategory(ctx context.Context, nama string) (int, error) {
	var n int
	err := (*d.stmt)[countProductByCategory].QueryRowxContext(ctx, nama).Scan(&n)
	return n, err
}

func (d Data) DeleteCategory(ctx context.Context, nama string) (int64, error) {
	r, err := (*d.stmt)[delCategory].ExecContext(ctx, nama)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

// ------------------------- PELANGGAN -------------------------

// FindCustomer mengembalikan sql.ErrNoRows bila belum ada.
func (d Data) FindCustomer(ctx context.Context, nama, telepon string) (int64, error) {
	var id int64
	err := (*d.stmt)[findCustomer].QueryRowxContext(ctx, nama, telepon).Scan(&id)
	return id, err
}

func (d Data) UpdateCustomerAlamat(ctx context.Context, id int64, alamat string) error {
	_, err := (*d.stmt)[updCustomerAlamat].ExecContext(ctx, alamat, id)
	return err
}

func (d Data) InsertCustomer(ctx context.Context, nama, telepon, alamat string) (int64, error) {
	r, err := (*d.stmt)[insCustomer].ExecContext(ctx, nama, nullStr(telepon), nullStr(alamat))
	if err != nil {
		return 0, err
	}
	return r.LastInsertId()
}

func (d Data) ListCustomers(ctx context.Context) ([]bakery.Customer, error) {
	rows, err := (*d.stmt)[listCustomers].QueryxContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.Customer{}
	for rows.Next() {
		var c bakery.Customer
		if err := rows.Scan(&c.ID, &c.Nama, &c.Telepon); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ------------------------- NOTIFIKASI & AUDIT -------------------------

func (d Data) ListNotifications(ctx context.Context) ([]bakery.Notification, error) {
	rows, err := (*d.stmt)[listNotifications].QueryxContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.Notification{}
	for rows.Next() {
		var n bakery.Notification
		var t time.Time
		if err := rows.Scan(&n.ID, &n.Pesan, &t); err != nil {
			return nil, err
		}
		n.Waktu = t.Format("2006-01-02 15:04")
		out = append(out, n)
	}
	return out, rows.Err()
}

func (d Data) MarkNotificationsRead(ctx context.Context) error {
	_, err := (*d.stmt)[readNotifications].ExecContext(ctx)
	return err
}

func (d Data) GetAuditLog(ctx context.Context, limit int) ([]bakery.AuditEntry, error) {
	rows, err := (*d.stmt)[getAuditLog].QueryxContext(ctx, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []bakery.AuditEntry{}
	for rows.Next() {
		var a bakery.AuditEntry
		var t time.Time
		var oid sql.NullInt64
		var kode, before, after sql.NullString
		if err := rows.Scan(&t, &oid, &kode, &a.Aksi, &a.Alasan, &a.Pelaku, &before, &after); err != nil {
			return nil, err
		}
		a.Waktu = t.Format("2006-01-02 15:04:05")
		a.Order = kode.String
		if oid.Valid {
			v := oid.Int64
			a.OrderID = &v
		}
		if before.Valid {
			a.Sebelum = json.RawMessage(before.String)
		}
		if after.Valid {
			a.Sesudah = json.RawMessage(after.String)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
