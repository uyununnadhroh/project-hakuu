package http

import (
	"errors"
	"log"
	"net/http"

	"bakery-scheduler/pkg/response"

	"github.com/gorilla/mux"
)

// Handler will initialize mux router and register handler
func (s *Server) Handler() *mux.Router {
	r := mux.NewRouter()
	// Jika tidak ditemukan, jangan diubah.
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	// Health Check
	r.HandleFunc("/", defaultHandler).Methods("GET")

	// Tambahan Prefix di depan API endpoint
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("", defaultHandler).Methods("GET")
	api.HandleFunc("/", defaultHandler).Methods("GET")
	api.HandleFunc("/BEversion", getVersion).Methods("GET")

	// Login (tanpa token)
	api.HandleFunc("/login", s.Bakery.Login).Methods("POST")

	// Semua endpoint di bawah ini wajib header Authorization: Bearer <token>
	p := api.NewRoute().Subrouter()
	p.Use(s.JWTMiddleware)

	// Akun
	p.HandleFunc("/users", s.Bakery.ListUsers).Methods("GET")

	// Order
	p.HandleFunc("/orders", s.Bakery.ListOrders).Methods("GET")
	p.HandleFunc("/orders", s.Bakery.CreateOrder).Methods("POST")
	p.HandleFunc("/orders/{id}", s.Bakery.GetOrder).Methods("GET")
	p.HandleFunc("/orders/{id}/{action}", s.Bakery.OrderAction).Methods("POST") // start, complete, cancel
	p.HandleFunc("/schedule", s.Bakery.Schedule).Methods("GET")

	// Negosiasi
	p.HandleFunc("/offers/{id}/offer", s.Bakery.OfferToBuyer).Methods("POST")
	p.HandleFunc("/offers/{id}/respond", s.Bakery.RespondOffer).Methods("POST")

	// Pengaturan
	p.HandleFunc("/settings", s.Bakery.GetSettings).Methods("GET")
	p.HandleFunc("/settings", s.Bakery.PutSettings).Methods("PUT")
	p.HandleFunc("/settings/weights", s.Bakery.PutWeights).Methods("PUT")

	// Audit & Notifikasi
	p.HandleFunc("/audit-log", s.Bakery.AuditLog).Methods("GET")
	p.HandleFunc("/notifications", s.Bakery.Notifications).Methods("GET")
	p.HandleFunc("/notifications/read", s.Bakery.ReadNotifications).Methods("POST")

	// Laporan
	p.HandleFunc("/reports/monthly", s.Bakery.MonthlyReport).Methods("GET")
	p.HandleFunc("/reports/monthly/export", s.Bakery.ExportMonthly).Methods("GET")
	p.HandleFunc("/reports/monthly/{month}/close", s.Bakery.CloseMonth).Methods("POST")

	// Master Produk
	p.HandleFunc("/products", s.Bakery.ListProducts).Methods("GET")
	p.HandleFunc("/products", s.Bakery.AddProduct).Methods("POST")
	p.HandleFunc("/products/{id}", s.Bakery.PatchProduct).Methods("PATCH")
	p.HandleFunc("/products/{id}", s.Bakery.DeleteProduct).Methods("DELETE")

	// Master Kategori
	p.HandleFunc("/categories", s.Bakery.ListCategories).Methods("GET")
	p.HandleFunc("/categories", s.Bakery.AddCategory).Methods("POST")
	p.HandleFunc("/categories/{nama}", s.Bakery.DeleteCategory).Methods("DELETE")

	// Master Pelanggan
	p.HandleFunc("/customers", s.Bakery.ListCustomers).Methods("GET")
	p.HandleFunc("/customers", s.Bakery.AddCustomer).Methods("POST")

	return r
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("bakery-scheduler Service API"))
}

func getVersion(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	resp.Data = map[string]string{"version": "BE:1.0"} // ubah versi di sini
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)

	err := errors.New("endpoint tidak ditemukan")
	log.Printf("[ERROR] %s %s - %v\n", r.Method, r.URL, err)
	resp.SetError(err, http.StatusNotFound)
}
