package auth

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service *Service
	issuer  *TokenIssuer
}

func NewHandler(service *Service, issuer *TokenIssuer) *Handler {
	return &Handler{service: service, issuer: issuer}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/signup", h.Signup)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgName     string `json:"org_name"`
		OrgSlug     string `json:"org_slug"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeHandlerError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	if req.Email == "" || req.Password == "" || req.OrgName == "" || req.OrgSlug == "" || req.DisplayName == "" {
		writeHandlerError(w, http.StatusBadRequest, "MISSING_FIELDS", "all fields are required")
		return
	}

	result, err := h.service.Signup(r.Context(), SignupRequest{
		OrgName:     req.OrgName,
		OrgSlug:     req.OrgSlug,
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		writeHandlerError(w, http.StatusBadRequest, "SIGNUP_FAILED", err.Error())
		return
	}

	writeHandlerJSON(w, http.StatusCreated, map[string]interface{}{
		"org_id":        result.OrgID,
		"user_id":       result.UserID,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeHandlerError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	result, err := h.service.Login(r.Context(), LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeHandlerError(w, http.StatusUnauthorized, "AUTH_FAILED", "invalid credentials")
		return
	}

	if result.MFARequired {
		writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
			"mfa_required": true,
			"user_id":      result.UserID,
		})
		return
	}

	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       result.UserID,
		"org_id":        result.OrgID,
		"role":          result.Role,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeHandlerError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	accessToken, refreshToken, err := h.service.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeHandlerError(w, http.StatusUnauthorized, "REFRESH_FAILED", "invalid or expired refresh token")
		return
	}

	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeHandlerError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	_ = h.service.Logout(r.Context(), req.RefreshToken)
	writeHandlerJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

// writeHandlerJSON writes a JSON response. Mirrors api.WriteJSON without the import cycle.
func writeHandlerJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeHandlerError writes a JSON error response. Mirrors api.WriteError without the import cycle.
func writeHandlerError(w http.ResponseWriter, status int, code, message string) {
	writeHandlerJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
