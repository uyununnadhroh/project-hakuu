package bakery

import (
	"net/http"

	"bakery-scheduler/pkg/response"

	"github.com/gorilla/mux"
)

// MonthlyReport GET /api/reports/monthly?month=2026-10
func (h *Handler) MonthlyReport(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	result, err := h.service.MonthlyReport(ctx, r.URL.Query().Get("month"))
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// CloseMonth POST /api/reports/monthly/{month}/close
func (h *Handler) CloseMonth(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	result, err := h.service.CloseMonth(ctx, mux.Vars(r)["month"])
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// ExportMonthly GET /api/reports/monthly/export?month=2026-10  (file CSV, bukan JSON)
func (h *Handler) ExportMonthly(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, month, err := h.service.ExportMonthlyCSV(ctx, r.URL.Query().Get("month"))
	if err != nil {
		resp := response.Response{}
		setError(&resp, r, err)
		resp.RenderJSON(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="laporan-`+month+`.csv"`)
	_, _ = w.Write(data)
	logInfo(r)
}
