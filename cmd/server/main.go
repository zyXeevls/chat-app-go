package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	httpHandler "github.com/zyXeevls/chat-app/internal/delivery/http"
	"github.com/zyXeevls/chat-app/internal/infrastructure/database"
	"github.com/zyXeevls/chat-app/internal/repository"
	"github.com/zyXeevls/chat-app/internal/websocket"
)

func main() {
	godotenv.Load()

	db := database.NewPostgres()
	defer db.Close()

	messageRepo := repository.NewMessageRepository(db)

	hub := websocket.NewHub(messageRepo)
	msgHandler := httpHandler.NewMessageHandler(db)

	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	http.HandleFunc("/messages", msgHandler.GetMessage)

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
