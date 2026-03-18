package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	httpHandler "github.com/zyXeevls/chat-app/internal/delivery/http"
	"github.com/zyXeevls/chat-app/internal/infrastructure/database"
	"github.com/zyXeevls/chat-app/internal/repository"
	"github.com/zyXeevls/chat-app/internal/usecase"
	"github.com/zyXeevls/chat-app/internal/websocket"
)

func main() {
	godotenv.Load()

	db := database.NewPostgres()
	defer db.Close()

	if err := database.EnsureSchema(db); err != nil {
		log.Fatalf("failed to ensure database schema: %v", err)
	}

	redisClient := database.NewRedis()
	defer redisClient.Close()

	messageRepo := repository.NewMessageRepository(db)
	authRepo := repository.NewAuthRepository(db)
	presenceRepo := &repository.PresenceRepository{Redis: redisClient}
	unreadRepo := &repository.UnreadRepository{DB: db}
	presenceUseCase := usecase.NewPresenceUseCase(presenceRepo)
	unreadUseCase := usecase.NewUnreadUseCase(unreadRepo)

	hub := websocket.NewHub(messageRepo, redisClient, presenceUseCase, unreadUseCase)
	msgHandler := httpHandler.NewMessageHandler(db)
	uploadHandler := httpHandler.NewUploadHandler()
	presenceHandler := httpHandler.NewPresenceHandler(presenceUseCase)
	unreadHandler := httpHandler.NewUnreadHandler(unreadUseCase)
	authHandler := httpHandler.NewAuthHandler(authRepo)
	fs := http.FileServer(http.Dir("./uploads"))

	go hub.Run()

	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})
	http.HandleFunc("/messages", msgHandler.GetMessage)
	http.HandleFunc("/upload", uploadHandler.UploadFile)
	http.HandleFunc("/presence", presenceHandler.GetStatus)
	http.HandleFunc("/unread", unreadHandler.GetUnread)
	http.HandleFunc("/unread/clear", unreadHandler.ClearUnread)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
