package api

import (
	"log/slog"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/opsnerve/fireline/internal/auth"
	"github.com/opsnerve/fireline/internal/messaging"
	"github.com/opsnerve/fireline/internal/tenant"
)

// MessagingHandler handles messaging API requests.
type MessagingHandler struct {
	svc *messaging.Service
}

// NewMessagingHandler creates a new MessagingHandler.
func NewMessagingHandler(svc *messaging.Service) *MessagingHandler {
	return &MessagingHandler{svc: svc}
}

// RegisterRoutes registers messaging API routes on the given mux.
func (h *MessagingHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	allRoles := requireRole("staff", "shift_manager", "gm", "ops_director", "owner")
	pinRoles := requireRole("shift_manager", "gm", "ops_director", "owner")

	mux.Handle("GET /api/v1/messaging/channels", chain(http.HandlerFunc(h.ListChannels), authMW, allRoles))
	mux.Handle("GET /api/v1/messaging/channels/{id}/messages", chain(http.HandlerFunc(h.ListMessages), authMW, allRoles))
	mux.Handle("POST /api/v1/messaging/channels/{id}/messages", chain(http.HandlerFunc(h.SendMessage), authMW, allRoles))
	mux.Handle("PUT /api/v1/messaging/messages/{id}/pin", chain(http.HandlerFunc(h.PinMessage), authMW, pinRoles))
}

// ListChannels returns channels for a location plus broadcast channels.
// GET /api/v1/messaging/channels?location_id=...
func (h *MessagingHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	locationID := r.URL.Query().Get("location_id")

	channels, err := h.svc.ListChannels(r.Context(), orgID, locationID)
	if err != nil {
		slog.Error("msg list channels error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "MSG_LIST_CHANNELS_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "channels", channels)
}

// ListMessages returns messages for a channel.
// GET /api/v1/messaging/channels/{id}/messages?limit=50
func (h *MessagingHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	channelID := r.PathValue("id")
	if channelID == "" {
		WriteError(w, http.StatusBadRequest, "MSG_MISSING_CHANNEL", "channel id is required")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	messages, err := h.svc.ListMessages(r.Context(), orgID, channelID, limit)
	if err != nil {
		slog.Error("msg list messages error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "MSG_LIST_MESSAGES_ERROR", "an internal error occurred")
		return
	}
	WriteList(w, http.StatusOK, "messages", messages)
}

// SendMessage sends a message to a channel.
// POST /api/v1/messaging/channels/{id}/messages
// Body: {"body": "..."}
func (h *MessagingHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	channelID := r.PathValue("id")
	if channelID == "" {
		WriteError(w, http.StatusBadRequest, "MSG_MISSING_CHANNEL", "channel id is required")
		return
	}

	userID := auth.UserIDFrom(r.Context())
	if userID == "" {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_USER", "no user context")
		return
	}

	var body struct {
		Body       string `json:"body"`
		SenderName string `json:"sender_name"`
		SenderRole string `json:"sender_role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if body.Body == "" {
		WriteError(w, http.StatusBadRequest, "MSG_MISSING_BODY", "message body is required")
		return
	}
	if body.SenderName == "" {
		WriteError(w, http.StatusBadRequest, "MSG_MISSING_SENDER_NAME", "sender_name is required")
		return
	}
	if body.SenderRole == "" {
		WriteError(w, http.StatusBadRequest, "MSG_MISSING_SENDER_ROLE", "sender_role is required")
		return
	}

	input := messaging.MessageInput{
		ChannelID:  channelID,
		SenderID:   userID,
		SenderName: body.SenderName,
		SenderRole: body.SenderRole,
		Body:       body.Body,
	}

	msg, err := h.svc.SendMessage(r.Context(), orgID, input)
	if err != nil {
		slog.Error("msg send error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "MSG_SEND_ERROR", "an internal error occurred")
		return
	}
	WriteJSON(w, http.StatusCreated, msg)
}

// PinMessage toggles the pinned status of a message.
// PUT /api/v1/messaging/messages/{id}/pin
// Body: {"pinned": true}
func (h *MessagingHandler) PinMessage(w http.ResponseWriter, r *http.Request) {
	orgID, err := tenant.OrgIDFrom(r.Context())
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "AUTH_NO_TENANT", "no tenant context")
		return
	}

	messageID := r.PathValue("id")
	if messageID == "" {
		WriteError(w, http.StatusBadRequest, "MSG_MISSING_MESSAGE", "message id is required")
		return
	}

	var body struct {
		Pinned bool `json:"pinned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	if err := h.svc.PinMessage(r.Context(), orgID, messageID, body.Pinned); err != nil {
		slog.Error("msg pin error", "error", err, "correlation_id", r.Header.Get("X-Request-ID"))
		WriteError(w, http.StatusInternalServerError, "MSG_PIN_ERROR", "an internal error occurred")
		return
	}

	status := "unpinned"
	if body.Pinned {
		status = "pinned"
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": status})
}
