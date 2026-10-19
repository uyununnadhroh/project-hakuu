package http

import (
	"net/http"

	"bakery-scheduler/pkg/grace"
)

type BakeryHandler interface {
	// auth
	Login(w http.ResponseWriter, r *http.Request)
	ListUsers(w http.ResponseWriter, r *http.Request)

	// order
	CreateOrder(w http.ResponseWriter, r *http.Request)
	ListOrders(w http.ResponseWriter, r *http.Request)
	GetOrder(w http.ResponseWriter, r *http.Request)
	OrderAction(w http.ResponseWriter, r *http.Request)
	Schedule(w http.ResponseWriter, r *http.Request)

	// negosiasi
	OfferToBuyer(w http.ResponseWriter, r *http.Request)
	RespondOffer(w http.ResponseWriter, r *http.Request)

	// pengaturan
	GetSettings(w http.ResponseWriter, r *http.Request)
	PutSettings(w http.ResponseWriter, r *http.Request)
	PutWeights(w http.ResponseWriter, r *http.Request)

	// audit & notifikasi
	AuditLog(w http.ResponseWriter, r *http.Request)
	Notifications(w http.ResponseWriter, r *http.Request)
	ReadNotifications(w http.ResponseWriter, r *http.Request)

	// laporan
	MonthlyReport(w http.ResponseWriter, r *http.Request)
	ExportMonthly(w http.ResponseWriter, r *http.Request)
	CloseMonth(w http.ResponseWriter, r *http.Request)

	// master produk, kategori, pelanggan
	ListProducts(w http.ResponseWriter, r *http.Request)
	AddProduct(w http.ResponseWriter, r *http.Request)
	PatchProduct(w http.ResponseWriter, r *http.Request)
	DeleteProduct(w http.ResponseWriter, r *http.Request)
	ListCategories(w http.ResponseWriter, r *http.Request)
	AddCategory(w http.ResponseWriter, r *http.Request)
	DeleteCategory(w http.ResponseWriter, r *http.Request)
	ListCustomers(w http.ResponseWriter, r *http.Request)
	AddCustomer(w http.ResponseWriter, r *http.Request)
}

// Server ...
type Server struct {
	Bakery         BakeryHandler
	JWTSecret      []byte
	AllowedOrigins []string // asal frontend yang diizinkan (CORS), mis. http://localhost:5173
}

// Serve is serving HTTP gracefully on port x ...
func (s *Server) Serve(port string) error {
	return grace.Serve(port, s.Routes())
}

// Routes router lengkap (dengan CORS). Dipisah dari Serve supaya bisa dites.
func (s *Server) Routes() http.Handler {
	return s.CORSMiddleware(s.Handler())
}
