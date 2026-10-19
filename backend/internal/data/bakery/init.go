package bakery

import (
	"context"
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
)

type (
	Data struct {
		db   *sqlx.DB
		stmt *map[string]*sqlx.Stmt
	}

	statement struct {
		key   string
		query string
	}
)

var (
	readStmt = []statement{
		// auth
		{getUserByEmail, qGetUserByEmail},
		{listUsers, qListUsers},

		// kapasitas
		{lockDay, qLockDay},
		{sumUsed, qSumUsed},
		{getDayCapacity, qGetDayCapacity},
		{sumUsedOvertime, qSumUsedOvertime},

		// order
		{getProductPrice, qGetProductPrice},
		{getCandidates, qGetCandidates},
		{getOrderStatusLock, qGetOrderStatusLock},
		{getOrderProdDay, qGetOrderProdDay},
		{getNewOrdersInWindow, qGetNewOrdersInWindow},
		{getOrdersForRescore, qGetOrdersForRescore},
		{getRebalanceItems, qGetRebalanceItems},
		{getRiskOrders, qGetRiskOrders},
		{listOrders, qListOrders},
		{getOrderByID, qGetOrderByID},
		{getOrderExtra, qGetOrderExtra},
		{getOrderItems, qGetOrderItems},
		{getOrderItemVariants, qGetOrderItemVariants},
		{getOrderOffers, qGetOrderOffers},
		{getOrderHistory, qGetOrderHistory},
		{getScheduleOrders, qGetScheduleOrders},

		// negosiasi
		{getOfferLock, qGetOfferLock},
		{countActiveOffers, qCountActiveOffers},
		{getExpiredOffers, qGetExpiredOffers},

		// pengaturan
		{getSettings, qGetSettings},
		{getActiveWeights, qGetActiveWeights},
		{countOverReserved, qCountOverReserved},

		// master
		{listProducts, qListProducts},
		{listProductVariants, qListProductVariants},
		{getVariantsByProduct, qGetVariantsByProduct},
		{getAllVariantsLock, qGetAllVariantsLock},
		{countVariantUsage, qCountVariantUsage},
		{countProductByID, qCountProductByID},
		{countProductUsage, qCountProductUsage},
		{listCategories, qListCategories},
		{countCategory, qCountCategory},
		{countProductByCategory, qCountProductByCategory},
		{findCustomer, qFindCustomer},
		{listCustomers, qListCustomers},
		{listNotifications, qListNotifications},
		{getAuditLog, qGetAuditLog},

		// laporan
		{getSummary, qGetSummary},
		{getDailyRevenue, qGetDailyRevenue},
		{getTopProducts, qGetTopProducts},
		{getTopCustomers, qGetTopCustomers},
		{getStatusCounts, qGetStatusCounts},
		{getDeliveredStats, qGetDeliveredStats},
		{getUtilization, qGetUtilization},
		{getOvertimeStats, qGetOvertimeStats},
		{getAuditCounts, qGetAuditCounts},
		{getMonthlyReport, qGetMonthlyReport},
		{countMonthlyReport, qCountMonthlyReport},
	}

	insertStmt = []statement{
		{insDayIgnore, qInsDayIgnore},
		{insReservation, qInsReservation},
		{insAudit, qInsAudit},
		{insNotification, qInsNotification},
		{insOrder, qInsOrder},
		{insOrderItem, qInsOrderItem},
		{insOrderItemVariant, qInsOrderItemVariant},
		{insOffer, qInsOffer},
		{upsertSetting, qUpsertSetting},
		{insWeights, qInsWeights},
		{insProduct, qInsProduct},
		{insProductVariant, qInsProductVariant},
		{insCategory, qInsCategory},
		{insCustomer, qInsCustomer},
		{insMonthlyReport, qInsMonthlyReport},
	}

	updateStmt = []statement{
		{updReservationDay, qUpdReservationDay},
		{lockDaysUntil, qLockDaysUntil},
		{updOrderKode, qUpdOrderKode},
		{updOrderNegotiation, qUpdOrderNegotiation},
		{updOrderProductionDay, qUpdOrderProductionDay},
		{updOrderAccepted, qUpdOrderAccepted},
		{updOrderStatus, qUpdOrderStatus},
		{updOrderDelivered, qUpdOrderDelivered},
		{updLateOrders, qUpdLateOrders},
		{updOrderScore, qUpdOrderScore},
		{updOfferOffered, qUpdOfferOffered},
		{updOfferAnswered, qUpdOfferAnswered},
		{updOrderAfterNego, qUpdOrderAfterNego},
		{expireRecommended, qExpireRecommended},
		{expireOpenByOrder, qExpireOpenByOrder},
		{updFutureCapacity, qUpdFutureCapacity},
		{deactivateWeights, qDeactivateWeights},
		{updProductNama, qUpdProductNama},
		{updProductHarga, qUpdProductHarga},
		{updProductAktif, qUpdProductAktif},
		{updProductKategori, qUpdProductKategori},
		{updProductJenisBobot, qUpdProductJenisBobot},
		{updProductBobot, qUpdProductBobot},
		{updProductVariant, qUpdProductVariant},
		{updProductMaksVarian, qUpdProductMaksVarian},
		{updCustomerAlamat, qUpdCustomerAlamat},
		{readNotifications, qReadNotifications},
	}

	deleteStmt = []statement{
		{delReservation, qDelReservation},
		{delReservationByOrder, qDelReservationByOrder},
		{delProduct, qDelProduct},
		{delProductVariant, qDelProductVariant},
		{delCategory, qDelCategory},
	}
)

// New ...
func New(db *sqlx.DB) *Data {
	var (
		stmts = make(map[string]*sqlx.Stmt)
	)
	d := &Data{
		db:   db,
		stmt: &stmts,
	}

	d.InitStmt()

	return d
}

func (d *Data) InitStmt() {
	var (
		err   error
		stmts = make(map[string]*sqlx.Stmt)
	)

	for _, v := range readStmt {
		stmts[v.key], err = d.db.PreparexContext(context.Background(), v.query)
		if err != nil {
			log.Fatalf("[DB] Failed to initialize select statement key %v, err : %v", v.key, err)
		}
	}

	for _, v := range insertStmt {
		stmts[v.key], err = d.db.PreparexContext(context.Background(), v.query)
		if err != nil {
			log.Fatalf("[DB] Failed to initialize insert statement key %v, err : %v", v.key, err)
		}
	}

	for _, v := range updateStmt {
		stmts[v.key], err = d.db.PreparexContext(context.Background(), v.query)
		if err != nil {
			log.Fatalf("[DB] Failed to initialize update statement key %v, err : %v", v.key, err)
		}
	}

	for _, v := range deleteStmt {
		stmts[v.key], err = d.db.PreparexContext(context.Background(), v.query)
		if err != nil {
			log.Fatalf("[DB] Failed to initialize delete statement key %v, err : %v", v.key, err)
		}
	}

	*d.stmt = stmts
}

// BeginTx membuka transaksi READ COMMITTED: setiap query melihat data terbaru
// yang sudah di-commit. Dikombinasikan dengan kunci baris daily_capacity
// (FOR UPDATE), pengecekan sisa kapasitas tidak bisa basi karena transaksi lain.
func (d Data) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return d.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
}

// st mengambil prepared statement. Bila tx tidak nil, statement dijalankan di
// dalam transaksi tersebut; bila nil, langsung ke database (autocommit).
func (d Data) st(ctx context.Context, tx *sqlx.Tx, key string) *sqlx.Stmt {
	s := (*d.stmt)[key]
	if tx != nil {
		return tx.StmtxContext(ctx, s)
	}
	return s
}

// nullID: id 0 disimpan sebagai NULL.
func nullID(id int64) interface{} {
	if id == 0 {
		return nil
	}
	return id
}

// nullStr: string kosong disimpan sebagai NULL.
func nullStr(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}
