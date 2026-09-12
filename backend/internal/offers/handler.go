package offers

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"kabadiconnect/backend/internal/httpx"
	"kabadiconnect/backend/internal/requestctx"
	"net/http"
)

type Handler struct {
	service *Service
	repo    *Repository
}

func NewHandler(s *Service, r *Repository) *Handler { return &Handler{service: s, repo: r} }
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	lotID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, 400, "INVALID_ID", "Lot id is invalid")
		return
	}
	var in CreateInput
	if !httpx.Decode(w, r, &in) {
		return
	}
	o, err := h.service.Create(r.Context(), requestctx.UserID(r.Context()), lotID, in)
	if err != nil {
		switch {
		case errors.Is(err, ErrQuantityUnavailable):
			httpx.Error(w, 409, "QUANTITY_UNAVAILABLE", err.Error())
		case errors.Is(err, ErrLotUnavailable):
			httpx.Error(w, 409, "LOT_UNAVAILABLE", err.Error())
		case errors.Is(err, ErrDuplicatePending):
			httpx.Error(w, 409, "PENDING_OFFER_EXISTS", err.Error())
		case errors.Is(err, ErrRecyclerUnauthorized):
			httpx.Error(w, 403, "RECYCLER_NOT_AUTHORIZED", err.Error())
		case errors.Is(err, ErrMaterialNotAccepted):
			httpx.Error(w, 403, "MATERIAL_NOT_ACCEPTED", err.Error())
		case errors.Is(err, mongo.ErrNoDocuments):
			httpx.Error(w, 404, "LOT_NOT_FOUND", "E-waste lot not found")
		default:
			httpx.Error(w, 400, "OFFER_FAILED", err.Error())
		}
		return
	}
	httpx.Data(w, 201, o)
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context(), requestctx.UserID(r.Context()), r.URL.Query().Get("status"))
	if err != nil {
		httpx.Error(w, 500, "OFFERS_FAILED", "Unable to load offers")
		return
	}
	httpx.Data(w, 200, map[string]any{"items": items})
}
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, 400, "INVALID_ID", "Offer id is invalid")
		return
	}
	if err := h.repo.CancelOwned(r.Context(), id, requestctx.UserID(r.Context())); err != nil {
		httpx.Error(w, 409, "OFFER_NOT_CANCELLABLE", "Only your pending offers can be cancelled")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
