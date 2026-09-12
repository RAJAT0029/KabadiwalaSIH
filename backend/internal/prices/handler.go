package prices

import (
	"kabadiconnect/backend/internal/httpx"
	"net/http"
	"strconv"
)

type Handler struct{ repo *Repository }

func NewHandler(r *Repository) *Handler { return &Handler{repo: r} }
func (h *Handler) Board(w http.ResponseWriter, r *http.Request) {
	b, err := h.repo.Board(r.Context(), r.URL.Query().Get("city"))
	if err != nil {
		httpx.Error(w, 500, "PRICES_FAILED", "Unable to load price board")
		return
	}
	httpx.Data(w, 200, b)
}
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	rows, err := h.repo.History(r.Context(), r.URL.Query().Get("category"), r.URL.Query().Get("subCategory"), limit)
	if err != nil {
		httpx.Error(w, 500, "PRICE_HISTORY_FAILED", "Unable to load price history")
		return
	}
	httpx.Data(w, 200, map[string]any{"items": rows})
}
