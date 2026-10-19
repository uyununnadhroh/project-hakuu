package bakery

// Query List to Prepare
const (
	// Pesanan dengan tanggal kirim hari ini, ditambah pesanan 7 hari terakhir yang
	// belum selesai (belum dikonfirmasi atau belum terkirim), beserta konfirmasi terakhirnya.
	getDeliveryOrders  = "GetDeliveryOrders"
	qGetDeliveryOrders = `SELECT o.id, COALESCE(o.kode,''), COALESCE(c.nama,''),
				COALESCE((SELECT GROUP_CONCAT(CONCAT(p.nama, ' x', i.jumlah) ORDER BY i.id SEPARATOR ', ')
					FROM order_items i JOIN products p ON p.id = i.product_id WHERE i.order_id = o.id), ''),
				o.waktu_kirim, o.status, o.diserahkan_pada,
				dc.status, dc.alasan, dc.dikonfirmasi_pada, u.nama
				FROM orders o
				LEFT JOIN customers c ON c.id = o.customer_id
				LEFT JOIN delivery_confirmations dc
					ON dc.id = (SELECT MAX(d2.id) FROM delivery_confirmations d2 WHERE d2.order_id = o.id)
				LEFT JOIN users u ON u.id = dc.dikonfirmasi_oleh
				WHERE o.status <> 'batal'
				  AND ((o.waktu_kirim >= ? AND o.waktu_kirim < ?)
				    OR (o.waktu_kirim >= ? AND o.waktu_kirim < ? AND o.status NOT IN ('selesai','baru','menunggu_negosiasi')))
				ORDER BY o.waktu_kirim, o.id`

	getOrderDeliveryLock  = "GetOrderDeliveryLock"
	qGetOrderDeliveryLock = `SELECT status, waktu_kirim FROM orders WHERE id = ? FOR UPDATE`

	insDeliveryConfirmation  = "InsDeliveryConfirmation"
	qInsDeliveryConfirmation = `INSERT INTO delivery_confirmations (order_id, status, alasan, dikonfirmasi_oleh, dikonfirmasi_pada)
				VALUES (?,?,?,?,?)`
)
