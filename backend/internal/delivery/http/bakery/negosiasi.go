package bakery

import (
	"net/http"
	"time"

	"bakery-scheduler/internal/entity/bakery"
	"bakery-scheduler/pkg/response"
)

// OfferToBuyer POST /api/offers/{id}/offer
func (h *Handler) OfferToBuyer(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	id, err := pathID(r)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	due, err := h.service.OfferToBuyer(ctx, id)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = map[string]string{"respond_before": due.Format(time.RFC3339)}
	logInfo(r)
}

// RespondOffer POST /api/offers/{id}/respond  body {"accept": true}
func (h *Handler) RespondOffer(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	id, err := pathID(r)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	var req bakery.ReqRespondOffer
	if err := decode(r, &req); err != nil {
		setError(&resp, r, err)
		return
	}

	if err := h.service.RespondOffer(ctx, id, req.Accept); err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = ok()
	logInfo(r)
}
