package transactions

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"kabadiconnect/backend/internal/httpx"
	"kabadiconnect/backend/internal/requestctx"
	"net/http"
)

type Handler struct{ repo *Repository }

func NewHandler(r *Repository) *Handler { return &Handler{repo: r} }
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListByRecycler(r.Context(), requestctx.UserID(r.Context()), r.URL.Query().Get("status"))
	if err != nil {
		httpx.Error(w, 500, "TRANSACTIONS_FAILED", "Unable to load transactions")
		return
	}
	httpx.Data(w, 200, map[string]any{"items": rows})
}
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, 400, "INVALID_ID", "Transaction id is invalid")
		return
	}
	row, err := h.repo.ByIDForRecycler(r.Context(), id, requestctx.UserID(r.Context()))
	if err != nil {
		httpx.Error(w, 404, "TRANSACTION_NOT_FOUND", "Transaction not found")
		return
	}
	httpx.Data(w, 200, row)
}
