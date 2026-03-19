package auth

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
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
	mux.HandleFunc("POST /api/v1/auth/pin-verify", h.PINVerify)
}

var pinAttempts sync.Map

type attemptTracker struct {
	mu       sync.Mutex
	failures int
	lockedAt time.Time
}

func checkPINLockout(locationID string) bool {
	val, ok := pinAttempts.Load(locationID)
	if !ok {
		return false
	}
	t := val.(*attemptTracker)
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.failures >= 5 && time.Since(t.lockedAt) < 5*time.Minute {
		return true
	}
	if time.Since(t.lockedAt) >= 5*time.Minute {
		t.failures = 0
	}
	return false
}

func recordPINFailure(locationID string) {
	val, _ := pinAttempts.LoadOrStore(locationID, &attemptTracker{})
	t := val.(*attemptTracker)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.failures++
	if t.failures >= 5 {
		t.lockedAt = time.Now()
	}
}

func resetPINAttempts(locationID string) {
	pinAttempts.Delete(locationID)
}

func (h *Handler) PINVerify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PIN        string `json:"pin"`
		LocationID string `json:"location_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeHandlerError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.PIN == "" || req.LocationID == "" {
		writeHandlerError(w, http.StatusBadRequest, "MISSING_FIELDS", "pin and location_id required")
		return
	}

	if checkPINLockout(req.LocationID) {
		writeHandlerError(w, http.StatusTooManyRequests, "PIN_LOCKED", "too many attempts, try again in 5 minutes")
		return
	}

	result, err := h.service.PINLogin(r.Context(), PINLoginRequest{
		LocationID: req.LocationID,
		PIN:        req.PIN,
	})
	if err != nil {
		recordPINFailure(req.LocationID)
		writeHandlerError(w, http.StatusUnauthorized, "PIN_FAILED", "invalid PIN")
		return
	}

	resetPINAttempts(req.LocationID)
	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
		"employee_id":  result.UserID,
		"display_name": result.DisplayName,
		"role":         result.Role,
	})
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
