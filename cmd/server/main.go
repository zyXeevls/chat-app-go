package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/zyXeevls/chat-app/internal/infrastructure/database"
	"github.com/zyXeevls/chat-app/internal/websocket"
)

func main() {
	godotenv.Load()

	db := database.NewPostgres()
	defer db.Close()

	hub := websocket.NewHub()

	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	log.Println("Server running on :8080")

	http.ListenAndServe(":8080", nil)
}
