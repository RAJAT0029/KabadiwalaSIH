package users

import (
	"net/http"

	"kabadiconnect/backend/internal/httpx"
	"kabadiconnect/backend/internal/requestctx"
)

type Handler struct{ repo *Repository }

func NewHandler(r *Repository) *Handler { return &Handler{repo: r} }

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u, err := h.repo.ByID(r.Context(), requestctx.UserID(r.Context()))
	if err != nil {
		httpx.Error(w, 404, "USER_NOT_FOUND", "Account not found")
		return
	}
	httpx.Data(w, 200, u)
}

func (h *Handler) PatchMe(w http.ResponseWriter, r *http.Request) {
	var in ProfilePatch
	if !httpx.Decode(w, r, &in) {
		return
	}
	u, err := h.repo.UpdateRecyclerProfile(r.Context(), requestctx.UserID(r.Context()), in)
	if err != nil {
		httpx.Error(w, 400, "PROFILE_UPDATE_FAILED", "Unable to update recycler profile")
		return
	}
	httpx.Data(w, 200, u)
}
