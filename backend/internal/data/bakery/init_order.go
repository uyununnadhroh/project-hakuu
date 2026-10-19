package bakery

// Kolom standar untuk daftar order (dipakai beberapa query di bawah).
const orderCols = `o.id, COALESCE(o.kode,''), c.nama, o.jumlah, o.qty_box,
				(SELECT GROUP_CONCAT(CONCAT(p.nama, ' x', i.jumlah) ORDER BY i.id SEPARATOR ', ')
					FROM order_items i JOIN products p ON p.id = i.product_id WHERE i.order_id = o.id),
				o.waktu_pesan, o.waktu_kirim, DATE_FORMAT(o.hari_produksi,'%Y-%m-%d'),
				o.status, o.skor_total, o.lembur_pcs,
				COALESCE(u.nama,''), COALESCE(u.email,'')`

// Query List to Prepare
const (
	getProductPrice  = "GetProductPrice"
	qGetProductPrice = `SELECT harga, aktif, bobot FROM products WHERE id = ?`

	insOrder  = "InsOrder"
	qInsOrder = `INSERT INTO orders (customer_id, jumlah, qty_box, total_harga, waktu_pesan, waktu_kirim, waktu_kirim_awal,
					hari_deadline, status, catatan, created_by)
				VALUES (?,?,?,?,?,?,?,?,'baru',?,?)`

	updOrderKode  = "UpdOrderKode"
	qUpdOrderKode = `UPDATE orders SET kode = ? WHERE id = ?`

	insOrderItem  = "InsOrderItem"
	qInsOrderItem = `INSERT INTO order_items (order_id, product_id, jumlah, harga_satuan, subtotal) VALUES (?,?,?,?,?)`

	updOrderNegotiation  = "UpdOrderNegotiation"
	qUpdOrderNegotiation = `UPDATE orders SET status = 'menunggu_negosiasi', skor_j = ?, skor_q = ?, skor_w = ?,
					skor_total = ?, skor_dihitung_pada = ?
				WHERE id = ?`

	updOrderProductionDay  = "UpdOrderProductionDay"
	qUpdOrderProductionDay = `UPDATE orders SET hari_produksi = ? WHERE id = ?`

	updOrderAccepted  = "UpdOrderAccepted"
	qUpdOrderAccepted = `UPDATE orders SET status = ?, hari_produksi = ?, lembur_pcs = ?, skor_j = ?, skor_q = ?, skor_w = ?,
					skor_total = ?, skor_dihitung_pada = ?
				WHERE id = ?`

	// order lain di hari yang sama yang masih boleh digeser
	getCandidates  = "GetCandidates"
	qGetCandidates = `SELECT r.id, r.order_id, o.kode, r.jumlah, o.waktu_kirim, o.waktu_pesan, DATE_FORMAT(o.hari_deadline,'%Y-%m-%d')
				FROM capacity_reservations r JOIN orders o ON o.id = r.order_id
				WHERE r.tanggal = ? AND r.tipe = 'tetap' AND r.lembur = 0 AND o.status IN ('baru','terjadwal')`

	getOrderStatusLock  = "GetOrderStatusLock"
	qGetOrderStatusLock = `SELECT status FROM orders WHERE id = ? FOR UPDATE`

	getOrderProdDay  = "GetOrderProdDay"
	qGetOrderProdDay = `SELECT DATE_FORMAT(hari_produksi,'%Y-%m-%d') FROM orders WHERE id = ?`

	updOrderStatus  = "UpdOrderStatus"
	qUpdOrderStatus = `UPDATE orders SET status = ? WHERE id = ?`

	updOrderDelivered  = "UpdOrderDelivered"
	qUpdOrderDelivered = `UPDATE orders SET diserahkan_pada = ? WHERE id = ?`

	getNewOrdersInWindow  = "GetNewOrdersInWindow"
	qGetNewOrdersInWindow = `SELECT id FROM orders WHERE status = 'baru' AND hari_produksi IS NOT NULL AND waktu_kirim <= ?`

	updLateOrders  = "UpdLateOrders"
	qUpdLateOrders = `UPDATE orders SET status = 'terlambat'
				WHERE status IN ('baru','terjadwal','at_risk','diproduksi') AND waktu_kirim < ?`

	getOrdersForRescore  = "GetOrdersForRescore"
	qGetOrdersForRescore = `SELECT id, jumlah, waktu_kirim, waktu_pesan FROM orders
				WHERE status IN ('baru','terjadwal','menunggu_negosiasi','at_risk')`

	updOrderScore  = "UpdOrderScore"
	qUpdOrderScore = `UPDATE orders SET skor_j = ?, skor_q = ?, skor_w = ?, skor_total = ?, skor_dihitung_pada = ? WHERE id = ?`

	getRebalanceItems  = "GetRebalanceItems"
	qGetRebalanceItems = `SELECT r.id, o.id, o.kode, r.jumlah, o.waktu_kirim, o.waktu_pesan,
					DATE_FORMAT(o.hari_deadline,'%Y-%m-%d'), DATE_FORMAT(r.tanggal,'%Y-%m-%d')
				FROM capacity_reservations r JOIN orders o ON o.id = r.order_id
				WHERE o.status = 'terjadwal' AND r.tipe = 'tetap' AND r.lembur = 0 AND r.tanggal > ?`

	getRiskOrders  = "GetRiskOrders"
	qGetRiskOrders = `SELECT id, kode FROM orders WHERE status = 'terjadwal' AND hari_produksi <= ?`

	// status kosong = semua status
	listOrders  = "ListOrders"
	qListOrders = `SELECT ` + orderCols + `
				FROM orders o LEFT JOIN customers c ON c.id = o.customer_id
				LEFT JOIN users u ON u.id = o.created_by
				WHERE ( o.status = ? OR '' = ? )
				ORDER BY o.waktu_kirim ASC
				LIMIT ?`

	getOrderByID  = "GetOrderByID"
	qGetOrderByID = `SELECT ` + orderCols + `
				FROM orders o LEFT JOIN customers c ON c.id = o.customer_id
				LEFT JOIN users u ON u.id = o.created_by
				WHERE o.id = ?`

	getOrderExtra  = "GetOrderExtra"
	qGetOrderExtra = `SELECT o.total_harga, o.skor_j, o.skor_q, o.skor_w, c.telepon, c.alamat, o.catatan
				FROM orders o LEFT JOIN customers c ON c.id = o.customer_id
				WHERE o.id = ?`

	getOrderItems  = "GetOrderItems"
	qGetOrderItems = `SELECT p.nama, i.jumlah, i.harga_satuan, i.subtotal
				FROM order_items i JOIN products p ON p.id = i.product_id
				WHERE i.order_id = ?`

	getOrderOffers  = "GetOrderOffers"
	qGetOrderOffers = `SELECT id, urutan, jam_alternatif, DATE_FORMAT(hari_produksi,'%Y-%m-%d'), status, batas_respons
				FROM negotiation_offers
				WHERE order_id = ?
				ORDER BY urutan`

	getOrderHistory  = "GetOrderHistory"
	qGetOrderHistory = `SELECT waktu, aksi, COALESCE(alasan,''), pelaku FROM audit_log WHERE order_id = ? ORDER BY id`

	getScheduleOrders  = "GetScheduleOrders"
	qGetScheduleOrders = `SELECT ` + orderCols + `
				FROM capacity_reservations r JOIN orders o ON o.id = r.order_id
				LEFT JOIN customers c ON c.id = o.customer_id
				LEFT JOIN users u ON u.id = o.created_by
				WHERE r.tanggal = ? AND r.tipe = 'tetap'
				ORDER BY o.skor_total DESC`
)
