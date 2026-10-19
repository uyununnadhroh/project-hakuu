package bakery

// Query List to Prepare
const (
	getSummary  = "GetSummary"
	qGetSummary = `SELECT COUNT(*), COALESCE(SUM(jumlah),0), COALESCE(SUM(total_harga),0)
				FROM orders
				WHERE status = 'selesai' AND waktu_kirim >= ? AND waktu_kirim < ?`

	getDailyRevenue  = "GetDailyRevenue"
	qGetDailyRevenue = `SELECT DATE_FORMAT(waktu_kirim,'%Y-%m-%d'), COUNT(*), SUM(jumlah), SUM(total_harga)
				FROM orders
				WHERE status = 'selesai' AND waktu_kirim >= ? AND waktu_kirim < ?
				GROUP BY 1 ORDER BY 1`

	getTopProducts  = "GetTopProducts"
	qGetTopProducts = `SELECT p.nama, SUM(i.jumlah), SUM(i.subtotal)
				FROM order_items i JOIN orders o ON o.id = i.order_id JOIN products p ON p.id = i.product_id
				WHERE o.status = 'selesai' AND o.waktu_kirim >= ? AND o.waktu_kirim < ?
				GROUP BY p.id, p.nama
				ORDER BY SUM(i.subtotal) DESC
				LIMIT 10`

	getTopCustomers  = "GetTopCustomers"
	qGetTopCustomers = `SELECT COALESCE(c.nama,'(tanpa nama)'), COUNT(*), SUM(o.total_harga)
				FROM orders o LEFT JOIN customers c ON c.id = o.customer_id
				WHERE o.status = 'selesai' AND o.waktu_kirim >= ? AND o.waktu_kirim < ?
				GROUP BY c.id, c.nama
				ORDER BY SUM(o.total_harga) DESC
				LIMIT 10`

	getStatusCounts  = "GetStatusCounts"
	qGetStatusCounts = `SELECT status, COUNT(*) FROM orders WHERE waktu_kirim >= ? AND waktu_kirim < ? GROUP BY status`

	getDeliveredStats  = "GetDeliveredStats"
	qGetDeliveredStats = `SELECT COUNT(*), COALESCE(SUM(diserahkan_pada <= waktu_kirim),0)
				FROM orders
				WHERE status = 'selesai' AND diserahkan_pada IS NOT NULL AND waktu_kirim >= ? AND waktu_kirim < ?`

	getUtilization  = "GetUtilization"
	qGetUtilization = `SELECT AVG(u.used / d.kapasitas_dasar) * 100
				FROM daily_capacity d
				JOIN (SELECT tanggal, SUM(jumlah) used FROM capacity_reservations WHERE tipe = 'tetap' GROUP BY tanggal) u
					ON u.tanggal = d.tanggal
				WHERE d.tanggal >= ? AND d.tanggal < ?`

	getOvertimeStats  = "GetOvertimeStats"
	qGetOvertimeStats = `SELECT COUNT(DISTINCT tanggal), COALESCE(SUM(lembur),0)
				FROM capacity_reservations
				WHERE lembur > 0 AND tanggal >= ? AND tanggal < ?`

	getAuditCounts  = "GetAuditCounts"
	qGetAuditCounts = `SELECT aksi, COUNT(*) FROM audit_log
				WHERE waktu >= ? AND waktu < ?
					AND aksi IN ('digeser','dimajukan','negosiasi_ditawarkan','negosiasi_diterima',
								'negosiasi_ditolak','negosiasi_kedaluwarsa','order_ditolak')
				GROUP BY aksi`

	getMonthlyReport  = "GetMonthlyReport"
	qGetMonthlyReport = `SELECT data, dibuat_pada FROM monthly_reports WHERE bulan = ?`

	countMonthlyReport  = "CountMonthlyReport"
	qCountMonthlyReport = `SELECT COUNT(*) FROM monthly_reports WHERE bulan = ?`

	insMonthlyReport  = "InsMonthlyReport"
	qInsMonthlyReport = `INSERT INTO monthly_reports (bulan, data) VALUES (?,?)`
)
