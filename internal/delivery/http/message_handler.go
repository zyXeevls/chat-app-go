package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageHandler struct {
	DB *pgxpool.Pool
}

func NewMessageHandler(db *pgxpool.Pool) *MessageHandler {
	return &MessageHandler{DB: db}
}

func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		http.Error(w, "room_id is required", http.StatusBadRequest)
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page == 0 {
		page = 1
	}

	if limit == 0 {
		limit = 20
	}

	rows, err := h.DB.Query(context.Background(),
		`SELECT room_id, sender_id, content, created_at
		FROM messages 
		WHERE room_id = $1 
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3`,
		roomID, limit, (page-1)*limit,
	)

	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	messages := []map[string]interface{}{}

	for rows.Next() {
		var roomID, senderID, content string
		var createdAt time.Time

		err = rows.Scan(&roomID, &senderID, &content, &createdAt)
		if err != nil {
			http.Error(w, "Failed to parse messages", http.StatusInternalServerError)
			return
		}

		messages = append(messages, map[string]interface{}{
			"room_id":    roomID,
			"sender_id":  senderID,
			"content":    content,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(messages)
}
