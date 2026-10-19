package bakery

import (
	"context"
	"database/sql"
	"time"

	"bakery-scheduler/internal/entity/bakery"

	"github.com/jmoiron/sqlx"
)

// BakeryData semua fungsi yang disediakan data layer. Parameter tx boleh nil
// (query langsung tanpa transaksi).
type BakeryData interface {
	BeginTx(ctx context.Context) (*sqlx.Tx, error)

	// auth
	GetUserByEmail(ctx context.Context, email string) (bakery.User, error)
	ListUsers(ctx context.Context) ([]bakery.UserView, error)

	// kapasitas, audit, notifikasi
	LockDay(ctx context.Context, tx *sqlx.Tx, day string) (int, int, bool, error)
	InsertDayIgnore(ctx context.Context, tx *sqlx.Tx, day string, base, overtimeMax int) error
	SumUsed(ctx context.Context, tx *sqlx.Tx, day string, now time.Time) (int, error)
	GetDayCapacity(ctx context.Context, day string) (int, bool, error)
	SumUsedOvertime(ctx context.Context, day string, now time.Time) (int, int, error)
	InsertReservation(ctx context.Context, tx *sqlx.Tx, orderID int64, day string, qty, overtime int, tipe string, expires *time.Time) (int64, error)
	UpdateReservationDay(ctx context.Context, tx *sqlx.Tx, resID int64, day string) error
	DeleteReservation(ctx context.Context, tx *sqlx.Tx, resID int64) error
	DeleteReservationsByOrder(ctx context.Context, tx *sqlx.Tx, orderID int64) error
	LockDaysUntil(ctx context.Context, tx *sqlx.Tx, day string) error
	InsertAudit(ctx context.Context, tx *sqlx.Tx, orderID int64, aksi string, before, after *string, alasan, pelaku string) error
	InsertNotification(ctx context.Context, tx *sqlx.Tx, orderID int64, pesan string) error

	// order
	GetProductPrice(ctx context.Context, tx *sqlx.Tx, productID int64) (bakery.ProductPrice, error)
	InsertOrder(ctx context.Context, tx *sqlx.Tx, o bakery.OrderInsert) (int64, error)
	UpdateOrderKode(ctx context.Context, tx *sqlx.Tx, id int64, kode string) error
	InsertOrderItem(ctx context.Context, tx *sqlx.Tx, orderID, productID int64, qty int, price float64) (int64, error)
	InsertOrderItemVariant(ctx context.Context, tx *sqlx.Tx, orderItemID, variantID int64, opsi string) error
	GetVariantsByProduct(ctx context.Context, tx *sqlx.Tx, productID int64) ([]bakery.ProductVariant, error)
	UpdateOrderNegotiation(ctx context.Context, tx *sqlx.Tx, id int64, sc bakery.Score, now time.Time) error
	UpdateOrderProductionDay(ctx context.Context, tx *sqlx.Tx, id int64, day string) error
	UpdateOrderAccepted(ctx context.Context, tx *sqlx.Tx, id int64, status, day string, overtime int, sc bakery.Score, now time.Time) error
	GetCandidates(ctx context.Context, tx *sqlx.Tx, day string) ([]bakery.Candidate, error)
	GetOrderStatusLock(ctx context.Context, tx *sqlx.Tx, id int64) (string, error)
	GetOrderProductionDay(ctx context.Context, tx *sqlx.Tx, id int64) (sql.NullString, error)
	UpdateOrderStatus(ctx context.Context, tx *sqlx.Tx, id int64, status string) error
	UpdateOrderDelivered(ctx context.Context, tx *sqlx.Tx, id int64, at time.Time) error
	GetDeliveryOrders(ctx context.Context, from, to, pastFrom time.Time) ([]bakery.DeliveryOrder, error)
	GetOrderDeliveryLock(ctx context.Context, tx *sqlx.Tx, id int64) (string, time.Time, error)
	InsertDeliveryConfirmation(ctx context.Context, tx *sqlx.Tx, orderID int64, status, alasan string, userID int64, at time.Time) error
	GetNewOrdersInWindow(ctx context.Context, tx *sqlx.Tx, until time.Time) ([]int64, error)
	MarkLateOrders(ctx context.Context, tx *sqlx.Tx, now time.Time) error
	GetOrdersForRescore(ctx context.Context, tx *sqlx.Tx) ([]bakery.ScoreRow, error)
	UpdateOrderScore(ctx context.Context, tx *sqlx.Tx, id int64, sc bakery.Score, now time.Time) error
	GetRebalanceItems(ctx context.Context, tx *sqlx.Tx, today string) ([]bakery.RebalanceItem, error)
	GetRiskOrders(ctx context.Context, tx *sqlx.Tx, today string) ([]bakery.RiskOrder, error)
	ListOrders(ctx context.Context, status string, limit int) ([]bakery.OrderRow, error)
	GetOrderByID(ctx context.Context, id int64) ([]bakery.OrderRow, error)
	GetScheduleOrders(ctx context.Context, day string) ([]bakery.OrderRow, error)
	GetOrderExtra(ctx context.Context, id int64) (bakery.OrderExtra, error)
	GetOrderItems(ctx context.Context, id int64) ([]bakery.OrderItemView, error)
	GetOrderOffers(ctx context.Context, id int64) ([]bakery.OrderOfferView, error)
	GetOrderHistory(ctx context.Context, id int64) ([]bakery.OrderHistoryView, error)

	// negosiasi
	InsertOffer(ctx context.Context, tx *sqlx.Tx, orderID int64, urutan int, jam time.Time, hariProduksi string) (int64, error)
	GetOfferLock(ctx context.Context, tx *sqlx.Tx, id int64) (bakery.OfferRow, error)
	CountActiveOffers(ctx context.Context, tx *sqlx.Tx, orderID int64, now time.Time) (int, error)
	UpdateOfferOffered(ctx context.Context, tx *sqlx.Tx, id int64, now, due time.Time, resID int64) error
	UpdateOfferAnswered(ctx context.Context, tx *sqlx.Tx, id int64, status string, now time.Time) error
	UpdateOrderAfterNegotiation(ctx context.Context, tx *sqlx.Tx, orderID int64, deliverAt time.Time, deadline, prodDay, status string, sc bakery.Score, now time.Time) error
	ExpireRecommendedOffers(ctx context.Context, tx *sqlx.Tx, orderID int64) error
	ExpireOpenOffersByOrder(ctx context.Context, tx *sqlx.Tx, orderID int64) error
	GetExpiredOffers(ctx context.Context, tx *sqlx.Tx, now time.Time) ([]bakery.ExpiredOffer, error)

	// pengaturan
	GetSettings(ctx context.Context, tx *sqlx.Tx) (map[string]float64, error)
	GetActiveWeights(ctx context.Context, tx *sqlx.Tx) (bakery.Weights, error)
	CountOverReserved(ctx context.Context, tx *sqlx.Tx, today string, now time.Time, limit float64) (int, error)
	UpdateFutureCapacity(ctx context.Context, tx *sqlx.Tx, base, overtimeMax int, today string) error
	UpsertSetting(ctx context.Context, tx *sqlx.Tx, key, value string) error
	DeactivateWeights(ctx context.Context, tx *sqlx.Tx) error
	InsertWeights(ctx context.Context, tx *sqlx.Tx, matrixJSON string, w []float64, cr float64) error

	// master
	ListProducts(ctx context.Context) ([]bakery.Product, error)
	ListProductVariants(ctx context.Context) ([]bakery.ProductVariant, error)
	CountProductByID(ctx context.Context, id int64) (int, error)
	InsertProduct(ctx context.Context, tx *sqlx.Tx, nama string, harga float64, kategori, jenis string, bobot float64, maksVarian int) (int64, error)
	GetAllVariantsLock(ctx context.Context, tx *sqlx.Tx, productID int64) ([]bakery.ProductVariant, error)
	CountVariantUsage(ctx context.Context, tx *sqlx.Tx, variantID int64) (int, error)
	InsertProductVariant(ctx context.Context, tx *sqlx.Tx, productID int64, v bakery.ProductVariant, urutan int) error
	UpdateProductVariant(ctx context.Context, tx *sqlx.Tx, productID int64, v bakery.ProductVariant, urutan int) error
	DeleteProductVariant(ctx context.Context, tx *sqlx.Tx, id int64) error
	UpdateProductMaksVarian(ctx context.Context, tx *sqlx.Tx, id int64, maks int) error
	UpdateProductNama(ctx context.Context, id int64, nama string) error
	UpdateProductHarga(ctx context.Context, id int64, harga float64) error
	UpdateProductAktif(ctx context.Context, id int64, aktif bool) error
	UpdateProductKategori(ctx context.Context, id int64, kategori string) error
	UpdateProductJenisBobot(ctx context.Context, id int64, jenis string, bobot float64) error
	UpdateProductBobot(ctx context.Context, id int64, bobot float64) error
	CountProductUsage(ctx context.Context, id int64) (int, error)
	DeleteProduct(ctx context.Context, id int64) (int64, error)
	ListCategories(ctx context.Context) ([]string, error)
	CountCategory(ctx context.Context, nama string) (int, error)
	InsertCategory(ctx context.Context, nama string) error
	CountProductByCategory(ctx context.Context, nama string) (int, error)
	DeleteCategory(ctx context.Context, nama string) (int64, error)
	FindCustomer(ctx context.Context, nama, telepon string) (int64, error)
	UpdateCustomerAlamat(ctx context.Context, id int64, alamat string) error
	InsertCustomer(ctx context.Context, nama, telepon, alamat string) (int64, error)
	ListCustomers(ctx context.Context) ([]bakery.Customer, error)
	ListNotifications(ctx context.Context) ([]bakery.Notification, error)
	MarkNotificationsRead(ctx context.Context) error
	GetAuditLog(ctx context.Context, limit int) ([]bakery.AuditEntry, error)

	// laporan
	GetSummary(ctx context.Context, from, to time.Time) (bakery.Summary, error)
	GetDailyRevenue(ctx context.Context, from, to time.Time) ([]bakery.DayRevenue, error)
	GetTopProducts(ctx context.Context, from, to time.Time) ([]bakery.ProductRow, error)
	GetTopCustomers(ctx context.Context, from, to time.Time) ([]bakery.CustomerRow, error)
	GetStatusCounts(ctx context.Context, from, to time.Time) (map[string]int, error)
	GetDeliveredStats(ctx context.Context, from, to time.Time) (int, int, error)
	GetUtilization(ctx context.Context, fromDay, toDay string) (sql.NullFloat64, error)
	GetOvertimeStats(ctx context.Context, fromDay, toDay string) (int, int, error)
	GetAuditCounts(ctx context.Context, from, to time.Time) ([]bakery.AuditCount, error)
	GetMonthlyReport(ctx context.Context, bulan string) (string, time.Time, error)
	CountMonthlyReport(ctx context.Context, bulan string) (int, error)
	InsertMonthlyReport(ctx context.Context, bulan, data string) error
}

type Service struct {
	bakery BakeryData
	loc    *time.Location
	secret []byte
	Now    func() time.Time // opsional, untuk pengujian
}

func New(data BakeryData, loc *time.Location, jwtSecret []byte) Service {
	return Service{
		bakery: data,
		loc:    loc,
		secret: jwtSecret,
	}
}
