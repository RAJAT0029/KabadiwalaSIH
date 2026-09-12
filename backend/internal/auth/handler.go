package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"kabadiconnect/backend/internal/config"
	"kabadiconnect/backend/internal/httpx"
)

type Handler struct {
	service *Service
	cfg     config.Config
}

func NewHandler(s *Service, c config.Config) *Handler { return &Handler{service: s, cfg: c} }

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in RegisterInput
	if !httpx.Decode(w, r, &in) {
		return
	}
	u, a, rt, err := h.service.Register(r.Context(), in)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			httpx.Error(w, 409, "ACCOUNT_EXISTS", "An account already uses that email or phone")
			return
		}
		httpx.Error(w, 400, "REGISTRATION_FAILED", err.Error())
		return
	}
	h.setRefresh(w, rt, h.cfg.RefreshTTL)
	httpx.Data(w, 201, map[string]any{"user": u, "accessToken": a})
}
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in loginRequest
	if !httpx.Decode(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Identifier) == "" || in.Password == "" {
		httpx.Error(w, 400, "VALIDATION_ERROR", "Email/phone and password are required")
		return
	}
	u, a, rt, err := h.service.Login(r.Context(), in.Identifier, in.Password, in.RememberMe)
	if err != nil {
		httpx.Error(w, 401, "INVALID_CREDENTIALS", "Email/phone or password is incorrect")
		return
	}
	ttl := h.cfg.RefreshTTL
	if !in.RememberMe {
		ttl = 24 * time.Hour
	}
	h.setRefresh(w, rt, ttl)
	httpx.Data(w, 200, map[string]any{"user": u, "accessToken": a})
}
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("kc_refresh")
	if err != nil {
		httpx.Error(w, 401, "UNAUTHENTICATED", "Refresh session is missing")
		return
	}
	a, rt, err := h.service.Refresh(r.Context(), c.Value)
	if err != nil {
		h.clearRefresh(w)
		httpx.Error(w, 401, "UNAUTHENTICATED", "Refresh session is invalid or expired")
		return
	}
	h.setRefresh(w, rt, h.cfg.RefreshTTL)
	httpx.Data(w, 200, map[string]any{"accessToken": a})
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("kc_refresh"); err == nil {
		h.service.Logout(r.Context(), c.Value)
	}
	h.clearRefresh(w)
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) setRefresh(w http.ResponseWriter, value string, ttl time.Duration) {
	if claims, err := h.service.parse(value, h.cfg.RefreshSecret, "refresh"); err == nil {
		ttl = time.Until(claims.ExpiresAt.Time)
	}
	http.SetCookie(w, &http.Cookie{Name: "kc_refresh", Value: value, Path: "/api/v1/auth", HttpOnly: true, Secure: h.cfg.Environment != "development", SameSite: http.SameSiteLaxMode, MaxAge: int(ttl.Seconds())})
}
func (h *Handler) clearRefresh(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "kc_refresh", Value: "", Path: "/api/v1/auth", HttpOnly: true, Secure: h.cfg.Environment != "development", SameSite: http.SameSiteLaxMode, MaxAge: -1})
}
