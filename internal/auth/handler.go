package auth

import (
	"encoding/json"
	"log/slog"
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
		"staff_points": result.StaffPoints,
		"points_trend": pinPointsTrend(result.StaffPoints),
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
		slog.Error("signup failed", "error", err, "email", req.Email)
		writeHandlerError(w, http.StatusBadRequest, "SIGNUP_FAILED", "unable to complete signup")
		return
	}

	setRefreshTokenCookie(w, r, result.RefreshToken)
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

	setRefreshTokenCookie(w, r, result.RefreshToken)
	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       result.UserID,
		"org_id":        result.OrgID,
		"role":          result.Role,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Prefer cookie, fall back to JSON body for mobile/tablet clients.
	refreshTokenIn := ""
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		refreshTokenIn = cookie.Value
	} else {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshTokenIn = req.RefreshToken
		}
	}

	if refreshTokenIn == "" {
		writeHandlerError(w, http.StatusBadRequest, "MISSING_TOKEN", "refresh token required")
		return
	}

	accessToken, refreshToken, err := h.service.RefreshAccessToken(r.Context(), refreshTokenIn)
	if err != nil {
		writeHandlerError(w, http.StatusUnauthorized, "REFRESH_FAILED", "invalid or expired refresh token")
		return
	}

	setRefreshTokenCookie(w, r, refreshToken)
	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Read refresh token from cookie or body to revoke server-side.
	refreshTokenIn := ""
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		refreshTokenIn = cookie.Value
	} else {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshTokenIn = req.RefreshToken
		}
	}

	if refreshTokenIn != "" {
		_ = h.service.Logout(r.Context(), refreshTokenIn)
	}

	// Clear the HttpOnly cookie regardless.
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth/refresh",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

// setRefreshTokenCookie attaches the refresh token as an HttpOnly cookie.
func setRefreshTokenCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth/refresh",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})
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

// pinPointsTrend returns a simple trend label for the PIN verify response.
// Because PIN login is pre-tenant (no DB access to history), we use the
// staff_points balance as a proxy: positive balance = up, zero = stable.
// Full trend computation (vs 7-day baseline) is available via the profile API.
func pinPointsTrend(staffPoints float64) string {
	switch {
	case staffPoints > 5:
		return "up"
	case staffPoints < -5:
		return "down"
	default:
		return "stable"
	}
}
