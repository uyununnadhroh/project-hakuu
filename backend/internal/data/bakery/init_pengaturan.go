package bakery

// Query List to Prepare
const (
	getSettings  = "GetSettings"
	qGetSettings = `SELECT kunci, nilai FROM settings`

	getActiveWeights  = "GetActiveWeights"
	qGetActiveWeights = `SELECT bobot_j, bobot_q, bobot_w FROM ahp_weights WHERE aktif = 1 ORDER BY id DESC LIMIT 1`

	// jumlah hari mendatang yang reservasinya melebihi kapasitas baru
	countOverReserved  = "CountOverReserved"
	qCountOverReserved = `SELECT COUNT(*) FROM (
					SELECT tanggal FROM capacity_reservations
					WHERE tanggal >= ? AND (tipe = 'tetap' OR kedaluwarsa_pada > ?)
					GROUP BY tanggal HAVING SUM(jumlah) > ?
				) x`

	updFutureCapacity  = "UpdFutureCapacity"
	qUpdFutureCapacity = `UPDATE daily_capacity SET kapasitas_dasar = ?, lembur_maks = ? WHERE tanggal >= ?`

	upsertSetting  = "UpsertSetting"
	qUpsertSetting = `INSERT INTO settings (kunci, nilai) VALUES (?,?)
				ON DUPLICATE KEY UPDATE nilai = VALUES(nilai), diubah_pada = CURRENT_TIMESTAMP`

	deactivateWeights  = "DeactivateWeights"
	qDeactivateWeights = `UPDATE ahp_weights SET aktif = 0`

	insWeights  = "InsWeights"
	qInsWeights = `INSERT INTO ahp_weights (matriks, bobot_j, bobot_q, bobot_w, cr, aktif) VALUES (?,?,?,?,?,1)`
)
