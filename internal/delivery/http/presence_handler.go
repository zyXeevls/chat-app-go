package http

import (
	"encoding/json"
	"net/http"

	"github.com/zyXeevls/chat-app/internal/usecase"
)

type PresenceHandler struct {
	usecase *usecase.PresenceUseCase
}

func NewPresenceHandler(presenceUseCase *usecase.PresenceUseCase) *PresenceHandler {
	return &PresenceHandler{usecase: presenceUseCase}
}

func (h *PresenceHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	status, lastSeen := h.usecase.GetStatus(userID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    status,
		"last_seen": lastSeen,
	})
}
