package listings

import (
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"kabadiconnect/backend/internal/httpx"
)

type Handler struct{ repo *Repository }

func NewHandler(r *Repository) *Handler { return &Handler{repo: r} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	qv := r.URL.Query()
	page, _ := strconv.Atoi(qv.Get("page"))
	limit, _ := strconv.Atoi(qv.Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	q := Query{Search: qv.Get("search"), Category: qv.Get("category"), SubCategory: firstNonEmpty(qv.Get("subCategory"), qv.Get("subType")), Condition: qv.Get("condition"), City: qv.Get("city"), Sort: qv.Get("sort"), MinQuantity: ParseFloat(qv.Get("minQuantity")), MaxPrice: ParseFloat(qv.Get("maxPrice")), Longitude: ParseFloat(qv.Get("longitude")), Latitude: ParseFloat(qv.Get("latitude")), Page: page, Limit: limit}
	items, total, err := h.repo.List(r.Context(), q)
	if err != nil {
		httpx.Error(w, 500, "LOTS_FAILED", "Unable to load e-waste lots")
		return
	}
	httpx.Data(w, 200, map[string]any{"items": items, "pagination": map[string]any{"page": page, "limit": limit, "total": total}})
}
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, 400, "INVALID_ID", "Lot id is invalid")
		return
	}
	q := r.URL.Query()
	v, err := h.repo.ByID(r.Context(), id, ParseFloat(q.Get("longitude")), ParseFloat(q.Get("latitude")))
	if err != nil {
		httpx.Error(w, 404, "LOT_NOT_FOUND", "E-waste lot not found")
		return
	}
	httpx.Data(w, 200, v)
}
func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
