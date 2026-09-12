package payments

import (
	"kabadiconnect/backend/internal/httpx"
	"kabadiconnect/backend/internal/requestctx"
	"net/http"
)

type Handler struct{ repo *Repository }

func NewHandler(r *Repository) *Handler { return &Handler{repo: r} }
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListByRecycler(r.Context(), requestctx.UserID(r.Context()))
	if err != nil {
		httpx.Error(w, 500, "PAYMENTS_FAILED", "Unable to load payments")
		return
	}
	httpx.Data(w, 200, map[string]any{"items": rows})
}
