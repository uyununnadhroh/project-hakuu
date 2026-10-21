package bakery

// Query List to Prepare
const (
	// ---------- produk ----------
	listProducts  = "ListProducts"
	qListProducts = `SELECT id, nama, harga, aktif, kategori, jenis, bobot, isi_box, isi_topping, maks_varian FROM products ORDER BY id`

	// semua varian (termasuk nonaktif, untuk Kelola Produk); Input Pesanan hanya memakai yang aktif
	listProductVariants  = "ListProductVariants"
	qListProductVariants = `SELECT id, product_id, nama, COALESCE(opsi_label,''), opsi_pilihan, aktif
				FROM product_variants ORDER BY product_id, urutan, id`

	getVariantsByProduct  = "GetVariantsByProduct"
	qGetVariantsByProduct = `SELECT id, product_id, nama, COALESCE(opsi_label,''), opsi_pilihan, aktif
				FROM product_variants WHERE product_id = ? AND aktif = TRUE ORDER BY urutan, id`

	getAllVariantsLock  = "GetAllVariantsLock"
	qGetAllVariantsLock = `SELECT id, product_id, nama, COALESCE(opsi_label,''), opsi_pilihan, aktif
				FROM product_variants WHERE product_id = ? ORDER BY urutan, id FOR UPDATE`

	countVariantUsage  = "CountVariantUsage"
	qCountVariantUsage = `SELECT COUNT(*) FROM order_item_variants WHERE variant_id = ?`

	insProductVariant  = "InsProductVariant"
	qInsProductVariant = `INSERT INTO product_variants (product_id, nama, opsi_label, opsi_pilihan, aktif, urutan) VALUES (?,?,?,?,?,?)`

	updProductVariant  = "UpdProductVariant"
	qUpdProductVariant = `UPDATE product_variants SET nama = ?, opsi_label = ?, opsi_pilihan = ?, aktif = ?, urutan = ?
				WHERE id = ? AND product_id = ?`

	delProductVariant  = "DelProductVariant"
	qDelProductVariant = `DELETE FROM product_variants WHERE id = ?`

	updProductMaksVarian  = "UpdProductMaksVarian"
	qUpdProductMaksVarian = `UPDATE products SET maks_varian = ? WHERE id = ?`

	countProductByID  = "CountProductByID"
	qCountProductByID = `SELECT COUNT(*) FROM products WHERE id = ?`

	insProduct  = "InsProduct"
	qInsProduct = `INSERT INTO products (nama, harga, kategori, jenis, bobot, isi_box, isi_topping, maks_varian) VALUES (?,?,?,?,?,?,?,?)`

	updProductNama  = "UpdProductNama"
	qUpdProductNama = `UPDATE products SET nama = ? WHERE id = ?`

	updProductHarga  = "UpdProductHarga"
	qUpdProductHarga = `UPDATE products SET harga = ? WHERE id = ?`

	updProductAktif  = "UpdProductAktif"
	qUpdProductAktif = `UPDATE products SET aktif = ? WHERE id = ?`

	updProductKategori  = "UpdProductKategori"
	qUpdProductKategori = `UPDATE products SET kategori = ? WHERE id = ?`

	getProductKomposisi  = "GetProductKomposisi"
	qGetProductKomposisi = `SELECT jenis, isi_box, isi_topping FROM products WHERE id = ?`

	updProductKomposisi  = "UpdProductKomposisi"
	qUpdProductKomposisi = `UPDATE products SET jenis = ?, isi_box = ?, isi_topping = ?, bobot = ? WHERE id = ?`

	updProductBobot  = "UpdProductBobot"
	qUpdProductBobot = `UPDATE products SET bobot = ? WHERE id = ?`

	countProductUsage  = "CountProductUsage"
	qCountProductUsage = `SELECT COUNT(*) FROM order_items WHERE product_id = ?`

	delProduct  = "DelProduct"
	qDelProduct = `DELETE FROM products WHERE id = ?`

	// ---------- kategori ----------
	listCategories  = "ListCategories"
	qListCategories = `SELECT nama FROM categories ORDER BY id`

	countCategory  = "CountCategory"
	qCountCategory = `SELECT COUNT(*) FROM categories WHERE nama = ?`

	insCategory  = "InsCategory"
	qInsCategory = `INSERT INTO categories (nama) VALUES (?)`

	countProductByCategory  = "CountProductByCategory"
	qCountProductByCategory = `SELECT COUNT(*) FROM products WHERE kategori = ?`

	delCategory  = "DelCategory"
	qDelCategory = `DELETE FROM categories WHERE nama = ?`

	// ---------- pelanggan ----------
	findCustomer  = "FindCustomer"
	qFindCustomer = `SELECT id FROM customers WHERE nama = ? AND COALESCE(telepon,'') = ? ORDER BY id LIMIT 1`

	updCustomerAlamat  = "UpdCustomerAlamat"
	qUpdCustomerAlamat = `UPDATE customers SET alamat = ? WHERE id = ?`

	insCustomer  = "InsCustomer"
	qInsCustomer = `INSERT INTO customers (nama, telepon, alamat) VALUES (?,?,?)`

	listCustomers  = "ListCustomers"
	qListCustomers = `SELECT id, nama, COALESCE(telepon,'') FROM customers ORDER BY nama`

	// ---------- notifikasi & audit ----------
	listNotifications  = "ListNotifications"
	qListNotifications = `SELECT id, pesan, dibuat_pada FROM notifications WHERE dibaca = 0 ORDER BY id DESC LIMIT 50`

	readNotifications  = "ReadNotifications"
	qReadNotifications = `UPDATE notifications SET dibaca = 1 WHERE dibaca = 0`

	getAuditLog  = "GetAuditLog"
	qGetAuditLog = `SELECT a.waktu, a.order_id, o.kode, a.aksi, COALESCE(a.alasan,''), a.pelaku, a.data_sebelum, a.data_sesudah
				FROM audit_log a LEFT JOIN orders o ON o.id = a.order_id
				ORDER BY a.id DESC
				LIMIT ?`
)
