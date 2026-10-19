package bakery

// Query List to Prepare
const (
	// ---------- produk ----------
	listProducts  = "ListProducts"
	qListProducts = `SELECT id, nama, harga, aktif, kategori, jenis, bobot FROM products ORDER BY id`

	countProductByID  = "CountProductByID"
	qCountProductByID = `SELECT COUNT(*) FROM products WHERE id = ?`

	insProduct  = "InsProduct"
	qInsProduct = `INSERT INTO products (nama, harga, kategori, jenis, bobot) VALUES (?,?,?,?,?)`

	updProductNama  = "UpdProductNama"
	qUpdProductNama = `UPDATE products SET nama = ? WHERE id = ?`

	updProductHarga  = "UpdProductHarga"
	qUpdProductHarga = `UPDATE products SET harga = ? WHERE id = ?`

	updProductAktif  = "UpdProductAktif"
	qUpdProductAktif = `UPDATE products SET aktif = ? WHERE id = ?`

	updProductKategori  = "UpdProductKategori"
	qUpdProductKategori = `UPDATE products SET kategori = ? WHERE id = ?`

	updProductJenisBobot  = "UpdProductJenisBobot"
	qUpdProductJenisBobot = `UPDATE products SET jenis = ?, bobot = ? WHERE id = ?`

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
