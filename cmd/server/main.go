package main

import (
	"log"
	"net/http"
	"time"

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
	rateLimiter := httpHandler.NewIPRateLimiter(60, time.Minute)

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/register", authHandler.Register)
	apiMux.HandleFunc("/login", authHandler.Login)
	apiMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})
	apiMux.HandleFunc("/messages", msgHandler.GetMessage)
	apiMux.HandleFunc("/upload", uploadHandler.UploadFile)
	apiMux.HandleFunc("/presence", presenceHandler.GetStatus)
	apiMux.HandleFunc("/unread", unreadHandler.GetUnread)
	apiMux.HandleFunc("/unread/clear", unreadHandler.ClearUnread)

	rootMux := http.NewServeMux()
	rootMux.Handle("/api/v1/", http.StripPrefix("/api/v1", rateLimiter.Middleware(apiMux)))
	rootMux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	go hub.Run()

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", rootMux))
}
