package bakery

import (
	"net/http"

	"bakery-scheduler/internal/entity/bakery"
	"bakery-scheduler/pkg/response"
)

// DeliveryConfirmations GET /api/delivery-confirmations
func (h *Handler) DeliveryConfirmations(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	result, err := h.service.DeliveryConfirmations(ctx)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// ConfirmDelivery POST /api/delivery-confirmations/{id}  (id = order)
func (h *Handler) ConfirmDelivery(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	id, err := pathID(r)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	var req bakery.ReqConfirmDelivery
	if err := decode(r, &req); err != nil {
		setError(&resp, r, err)
		return
	}

	if err := h.service.ConfirmDelivery(ctx, id, req, currentActor(r)); err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = ok()
	logInfo(r)
}
