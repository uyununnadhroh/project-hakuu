package bakery

import (
	"net/http"
	"strconv"

	"bakery-scheduler/internal/entity/bakery"
	"bakery-scheduler/pkg/response"

	"github.com/gorilla/mux"
)

// CreateOrder POST /api/orders
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	var req bakery.ReqCreateOrder
	if err := decode(r, &req); err != nil {
		setError(&resp, r, err)
		return
	}

	result, err := h.service.CreateOrder(ctx, req, currentActor(r))
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.StatusCode = http.StatusCreated
	switch result.Outcome {
	case bakery.OutcomeRejected:
		resp.StatusCode = http.StatusUnprocessableEntity
	case bakery.OutcomeCheckStock: // belum ada yang disimpan
		resp.StatusCode = http.StatusOK
	}
	resp.Data = result
	logInfo(r)
}

// ListOrders GET /api/orders?status=&limit=
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	result, err := h.service.ListOrders(ctx, r.URL.Query().Get("status"), limit)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// GetOrder GET /api/orders/{id}
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	id, err := pathID(r)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	result, err := h.service.GetOrder(ctx, id)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// OrderAction POST /api/orders/{id}/{action}  (action: start, complete, cancel)
func (h *Handler) OrderAction(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	id, err := pathID(r)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	if err := h.service.Transition(ctx, id, mux.Vars(r)["action"], currentActor(r).Nama); err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = ok()
	logInfo(r)
}

// Schedule GET /api/schedule?from=YYYY-MM-DD&days=7
func (h *Handler) Schedule(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	result, err := h.service.Schedule(ctx, r.URL.Query().Get("from"), days)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}
