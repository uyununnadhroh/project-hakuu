package bakery

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"bakery-scheduler/internal/entity"
	"bakery-scheduler/internal/entity/bakery"
	"bakery-scheduler/pkg/response"

	"github.com/gorilla/mux"
)

type Service interface {
	// auth
	Login(ctx context.Context, req bakery.ReqLogin) (bakery.LoginResult, error)
	ListUsers(ctx context.Context) ([]bakery.UserView, error)

	// order
	CreateOrder(ctx context.Context, req bakery.ReqCreateOrder, actor bakery.Actor) (*bakery.Result, error)
	ListOrders(ctx context.Context, status string, limit int) ([]bakery.OrderRow, error)
	GetOrder(ctx context.Context, id int64) (*bakery.OrderDetail, error)
	Transition(ctx context.Context, orderID int64, action, actor string) error
	Schedule(ctx context.Context, fromStr string, days int) ([]bakery.DayView, error)

	// negosiasi
	OfferToBuyer(ctx context.Context, offerID int64) (time.Time, error)
	RespondOffer(ctx context.Context, offerID int64, accept bool) error

	// pengaturan, audit, notifikasi
	GetSettingsView(ctx context.Context) (map[string]interface{}, error)
	UpdateSettings(ctx context.Context, in map[string]float64) error
	UpdateWeights(ctx context.Context, matrix [][]float64) (map[string]interface{}, bool, error)
	AuditLog(ctx context.Context, limit int) ([]bakery.AuditEntry, error)
	Notifications(ctx context.Context) ([]bakery.Notification, error)
	MarkNotificationsRead(ctx context.Context) error

	// laporan
	MonthlyReport(ctx context.Context, monthStr string) (*bakery.Report, error)
	CloseMonth(ctx context.Context, monthStr string) (*bakery.Report, error)
	ExportMonthlyCSV(ctx context.Context, month string) ([]byte, string, error)

	// master
	ListProducts(ctx context.Context) ([]bakery.Product, error)
	AddProduct(ctx context.Context, in bakery.ProductInput) (int64, error)
	PatchProduct(ctx context.Context, id int64, in bakery.ReqPatchProduct) error
	DeleteProduct(ctx context.Context, id int64) error
	ListCategories(ctx context.Context) ([]string, error)
	AddCategory(ctx context.Context, nama string) error
	DeleteCategory(ctx context.Context, nama string) error
	ListCustomers(ctx context.Context) ([]bakery.Customer, error)
	FindOrCreateCustomer(ctx context.Context, nama, telepon, alamat string) (int64, error)
}

type (
	Handler struct {
		service Service
	}
)

func New(s Service) *Handler {
	return &Handler{
		service: s,
	}
}

// ------------------------- helper -------------------------

// setError memetakan error service ke HTTP status lalu mencatat log.
// Detail error internal (500) tidak dikirim ke klien.
func setError(resp *response.Response, r *http.Request, err error) {
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, bakery.ErrLoginFailed):
		code = http.StatusUnauthorized
	case errors.Is(err, bakery.ErrValidation):
		code = http.StatusBadRequest
	case errors.Is(err, bakery.ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, bakery.ErrState), errors.Is(err, bakery.ErrExpired), errors.Is(err, bakery.ErrCapacity):
		code = http.StatusConflict
	}
	log.Printf("[ERROR][%s][%s] %s | Reason: %s", r.RemoteAddr, r.Method, r.URL, err.Error())
	if code == http.StatusInternalServerError {
		err = errors.New("terjadi kesalahan pada server") // detail internal tidak dibocorkan ke klien
	}
	resp.SetError(err, code)
}

func decode(r *http.Request, v interface{}) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errors.Join(bakery.ErrValidation, errors.New("body JSON tidak valid"))
	}
	return nil
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.Join(bakery.ErrValidation, errors.New("id tidak valid"))
	}
	return id, nil
}

func ok() map[string]string { return map[string]string{"status": "ok"} }

// currentActor membaca user yang login dari context (diisi JWTMiddleware).
func currentActor(r *http.Request) bakery.Actor {
	a := bakery.Actor{Nama: "penjual"}
	cv, okCtx := r.Context().Value(entity.ContextKey("claims")).(entity.ContextValue)
	if !okCtx {
		return a
	}
	if sub, isStr := cv.Get("sub").(string); isStr {
		if id, err := strconv.ParseInt(sub, 10, 64); err == nil {
			a.UserID = id
		}
	}
	if name, isStr := cv.Get("name").(string); isStr && name != "" {
		if rn := []rune(name); len(rn) > 50 { // kolom audit_log.pelaku VARCHAR(50)
			name = string(rn[:50])
		}
		a.Nama = name
	}
	return a
}

func logInfo(r *http.Request) {
	log.Printf("[INFO][%s][%s] %s", r.RemoteAddr, r.Method, r.URL)
}
