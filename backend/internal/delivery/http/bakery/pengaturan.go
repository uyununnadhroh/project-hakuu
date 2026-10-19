package bakery

import (
	"net/http"
	"strconv"

	"bakery-scheduler/internal/entity/bakery"
	"bakery-scheduler/pkg/response"
)

// GetSettings GET /api/settings
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	result, err := h.service.GetSettingsView(ctx)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// PutSettings PUT /api/settings  body {"kapasitas_harian": 800, ...}
func (h *Handler) PutSettings(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	var req map[string]float64
	if err := decode(r, &req); err != nil {
		setError(&resp, r, err)
		return
	}

	if err := h.service.UpdateSettings(ctx, req); err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = ok()
	logInfo(r)
}

// PutWeights PUT /api/settings/weights  body {"matrix": [[...],[...],[...]]}
func (h *Handler) PutWeights(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	var req bakery.ReqWeights
	if err := decode(r, &req); err != nil {
		setError(&resp, r, err)
		return
	}

	result, consistent, err := h.service.UpdateWeights(ctx, req.Matrix)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	if !consistent {
		resp.StatusCode = http.StatusUnprocessableEntity
	}
	resp.Data = result
	logInfo(r)
}

// AuditLog GET /api/audit-log?limit=
func (h *Handler) AuditLog(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	result, err := h.service.AuditLog(ctx, limit)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// Notifications GET /api/notifications
func (h *Handler) Notifications(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	result, err := h.service.Notifications(ctx)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// ReadNotifications POST /api/notifications/read
func (h *Handler) ReadNotifications(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	if err := h.service.MarkNotificationsRead(ctx); err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = ok()
	logInfo(r)
}
