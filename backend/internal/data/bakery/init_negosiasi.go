package bakery

// Query List to Prepare
const (
	insOffer  = "InsOffer"
	qInsOffer = `INSERT INTO negotiation_offers (order_id, urutan, jam_alternatif, hari_produksi) VALUES (?,?,?,?)`

	getOfferLock  = "GetOfferLock"
	qGetOfferLock = `SELECT f.id, f.order_id, f.status, f.jam_alternatif, DATE_FORMAT(f.hari_produksi,'%Y-%m-%d'),
					f.reservation_id, f.batas_respons, o.jumlah, o.status, o.kode, o.waktu_pesan
				FROM negotiation_offers f JOIN orders o ON o.id = f.order_id
				WHERE f.id = ?
				FOR UPDATE`

	countActiveOffers  = "CountActiveOffers"
	qCountActiveOffers = `SELECT COUNT(*) FROM negotiation_offers
				WHERE order_id = ? AND status = 'ditawarkan' AND batas_respons > ?`

	updOfferOffered  = "UpdOfferOffered"
	qUpdOfferOffered = `UPDATE negotiation_offers SET status = 'ditawarkan', ditawarkan_pada = ?, batas_respons = ?, reservation_id = ?
				WHERE id = ?`

	updOfferAnswered  = "UpdOfferAnswered"
	qUpdOfferAnswered = `UPDATE negotiation_offers SET status = ?, reservation_id = NULL, dijawab_pada = ? WHERE id = ?`

	updOrderAfterNego  = "UpdOrderAfterNego"
	qUpdOrderAfterNego = `UPDATE orders SET waktu_kirim = ?, hari_deadline = ?, hari_produksi = ?, status = ?,
					skor_j = ?, skor_q = ?, skor_w = ?, skor_total = ?, skor_dihitung_pada = ?
				WHERE id = ?`

	expireRecommended  = "ExpireRecommended"
	qExpireRecommended = `UPDATE negotiation_offers SET status = 'kedaluwarsa' WHERE order_id = ? AND status = 'direkomendasikan'`

	expireOpenByOrder  = "ExpireOpenByOrder"
	qExpireOpenByOrder = `UPDATE negotiation_offers SET status = 'kedaluwarsa', reservation_id = NULL
				WHERE order_id = ? AND status IN ('direkomendasikan','ditawarkan')`

	getExpiredOffers  = "GetExpiredOffers"
	qGetExpiredOffers = `SELECT id, order_id, reservation_id FROM negotiation_offers
				WHERE status = 'ditawarkan' AND batas_respons <= ?`
)
