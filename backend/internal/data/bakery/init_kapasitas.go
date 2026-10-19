package bakery

// Query List to Prepare
const (
	lockDay  = "LockDay"
	qLockDay = `SELECT kapasitas_dasar, lembur_maks, terkunci
				FROM daily_capacity
				WHERE tanggal = ?
				FOR UPDATE`

	insDayIgnore  = "InsDayIgnore"
	qInsDayIgnore = `INSERT IGNORE INTO daily_capacity (tanggal, kapasitas_dasar, lembur_maks) VALUES (?,?,?)`

	// reservasi sementara yang sudah kedaluwarsa tidak dihitung
	sumUsed  = "SumUsed"
	qSumUsed = `SELECT COALESCE(SUM(jumlah),0)
				FROM capacity_reservations
				WHERE tanggal = ? AND (tipe = 'tetap' OR kedaluwarsa_pada > ?)`

	getDayCapacity  = "GetDayCapacity"
	qGetDayCapacity = `SELECT kapasitas_dasar, terkunci FROM daily_capacity WHERE tanggal = ?`

	sumUsedOvertime  = "SumUsedOvertime"
	qSumUsedOvertime = `SELECT COALESCE(SUM(jumlah),0), COALESCE(SUM(lembur),0)
				FROM capacity_reservations
				WHERE tanggal = ? AND (tipe = 'tetap' OR kedaluwarsa_pada > ?)`

	insReservation  = "InsReservation"
	qInsReservation = `INSERT INTO capacity_reservations (order_id, tanggal, jumlah, lembur, tipe, kedaluwarsa_pada)
				VALUES (?,?,?,?,?,?)`

	updReservationDay  = "UpdReservationDay"
	qUpdReservationDay = `UPDATE capacity_reservations SET tanggal = ? WHERE id = ?`

	delReservation  = "DelReservation"
	qDelReservation = `DELETE FROM capacity_reservations WHERE id = ?`

	delReservationByOrder  = "DelReservationByOrder"
	qDelReservationByOrder = `DELETE FROM capacity_reservations WHERE order_id = ?`

	lockDaysUntil  = "LockDaysUntil"
	qLockDaysUntil = `UPDATE daily_capacity SET terkunci = 1 WHERE tanggal <= ?`

	insAudit  = "InsAudit"
	qInsAudit = `INSERT INTO audit_log (order_id, aksi, data_sebelum, data_sesudah, alasan, pelaku) VALUES (?,?,?,?,?,?)`

	insNotification  = "InsNotification"
	qInsNotification = `INSERT INTO notifications (order_id, pesan) VALUES (?,?)`
)
