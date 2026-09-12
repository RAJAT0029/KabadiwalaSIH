package handovers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"kabadiconnect/backend/internal/httpx"
	"kabadiconnect/backend/internal/requestctx"
	"net/http"
)

type Handler struct{ repo *Repository }

func NewHandler(r *Repository) *Handler { return &Handler{repo: r} }
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListByRecycler(r.Context(), requestctx.UserID(r.Context()))
	if err != nil {
		httpx.Error(w, 500, "HANDOVERS_FAILED", "Unable to load handover records")
		return
	}
	httpx.Data(w, 200, map[string]any{"items": rows})
}
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, 400, "INVALID_ID", "Handover id is invalid")
		return
	}
	row, err := h.repo.ByIDForRecycler(r.Context(), id, requestctx.UserID(r.Context()))
	if err != nil {
		httpx.Error(w, 404, "HANDOVER_NOT_FOUND", "Handover record not found")
		return
	}
	httpx.Data(w, 200, row)
}
func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, 400, "INVALID_ID", "Handover id is invalid")
		return
	}
	row, err := h.repo.ConfirmRecycler(r.Context(), id, requestctx.UserID(r.Context()))
	if err != nil {
		httpx.Error(w, 409, "HANDOVER_NOT_CONFIRMABLE", "This handover cannot be confirmed by the current recycler")
		return
	}
	httpx.Data(w, 200, row)
}
