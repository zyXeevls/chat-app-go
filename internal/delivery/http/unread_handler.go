package http

import (
	"encoding/json"
	"net/http"

	"github.com/zyXeevls/chat-app/internal/usecase"
)

type UnreadHandler struct {
	usecase *usecase.UnreadUseCase
}

func NewUnreadHandler(unreadUseCase *usecase.UnreadUseCase) *UnreadHandler {
	return &UnreadHandler{usecase: unreadUseCase}
}

func (h *UnreadHandler) GetUnread(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	result, err := h.usecase.GetUnread(userID)
	if err != nil {
		http.Error(w, "failed to get unread", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *UnreadHandler) ClearUnread(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		UserID string `json:"user_id"`
		RoomID string `json:"room_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.UserID == "" || body.RoomID == "" {
		http.Error(w, "user_id and room_id are required", http.StatusBadRequest)
		return
	}

	h.usecase.ClearUnread(body.UserID, body.RoomID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
