package bakery

import (
	"net/http"

	"bakery-scheduler/internal/entity/bakery"
	"bakery-scheduler/pkg/response"
)

// Login POST /api/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	var req bakery.ReqLogin
	if err := decode(r, &req); err != nil {
		setError(&resp, r, err)
		return
	}

	result, err := h.service.Login(ctx, req)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// ListUsers GET /api/users — email yang bisa masuk ke website
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	result, err := h.service.ListUsers(ctx)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}
