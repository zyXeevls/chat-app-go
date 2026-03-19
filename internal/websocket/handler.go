package websocket

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/zyXeevls/chat-app/pkg/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	userID := ""

	if token != "" {
		validatedUserID, err := utils.ValidateToken(token)
		if err != nil {
			http.Error(w, "unauthorized", 401)
			return
		}
		userID = validatedUserID
	} else {
		userID = strings.TrimSpace(r.URL.Query().Get("user_id"))
		if userID == "" {
			http.Error(w, "token or user_id is required", http.StatusBadRequest)
			return
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade connection", 500)
		return

	}

	client := &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan []byte, 256),
		userID:      userID,
		joinedRooms: make(map[string]bool),
	}

	client.hub.register <- client

	go client.writePump()

	go client.readPump()
}
