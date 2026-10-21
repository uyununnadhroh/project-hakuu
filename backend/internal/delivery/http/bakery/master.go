package bakery

import (
	"net/http"

	"bakery-scheduler/internal/entity/bakery"
	"bakery-scheduler/pkg/response"

	"github.com/gorilla/mux"
)

// ------------------------- PRODUK -------------------------

// ListProducts GET /api/products
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	result, err := h.service.ListProducts(ctx)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// AddProduct POST /api/products
func (h *Handler) AddProduct(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	var req bakery.ReqProduct
	if err := decode(r, &req); err != nil {
		setError(&resp, r, err)
		return
	}

	id, err := h.service.AddProduct(ctx, bakery.ProductInput{
		Nama: req.Nama, Harga: req.Harga, Kategori: req.Kategori, Jenis: req.Jenis, Bobot: req.Bobot,
		IsiBox: req.IsiBox, IsiTopping: req.IsiTopping,
		MaksVarian: req.MaksVarian, Variants: req.Variants})
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.StatusCode = http.StatusCreated
	resp.Data = map[string]int64{"id": id}
	logInfo(r)
}

// PatchProduct PATCH /api/products/{id}
func (h *Handler) PatchProduct(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	id, err := pathID(r)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	var req bakery.ReqPatchProduct
	if err := decode(r, &req); err != nil {
		setError(&resp, r, err)
		return
	}

	if err := h.service.PatchProduct(ctx, id, req); err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = ok()
	logInfo(r)
}

// DeleteProduct DELETE /api/products/{id}
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	id, err := pathID(r)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	if err := h.service.DeleteProduct(ctx, id); err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = ok()
	logInfo(r)
}

// ------------------------- KATEGORI -------------------------

// ListCategories GET /api/categories
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	result, err := h.service.ListCategories(ctx)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// AddCategory POST /api/categories  body {"nama": "..."}
func (h *Handler) AddCategory(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	var req bakery.ReqCategory
	if err := decode(r, &req); err != nil {
		setError(&resp, r, err)
		return
	}

	if err := h.service.AddCategory(ctx, req.Nama); err != nil {
		setError(&resp, r, err)
		return
	}

	resp.StatusCode = http.StatusCreated
	resp.Data = ok()
	logInfo(r)
}

// DeleteCategory DELETE /api/categories/{nama}
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	if err := h.service.DeleteCategory(ctx, mux.Vars(r)["nama"]); err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = ok()
	logInfo(r)
}

// ------------------------- PELANGGAN -------------------------

// ListCustomers GET /api/customers
func (h *Handler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	result, err := h.service.ListCustomers(ctx)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.Data = result
	logInfo(r)
}

// AddCustomer POST /api/customers  body {"nama","telepon","alamat"}
func (h *Handler) AddCustomer(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.RenderJSON(w, r)
	ctx := r.Context()

	var req bakery.ReqCustomer
	if err := decode(r, &req); err != nil {
		setError(&resp, r, err)
		return
	}

	id, err := h.service.FindOrCreateCustomer(ctx, req.Nama, req.Telepon, req.Alamat)
	if err != nil {
		setError(&resp, r, err)
		return
	}

	resp.StatusCode = http.StatusCreated
	resp.Data = map[string]int64{"id": id}
	logInfo(r)
}
