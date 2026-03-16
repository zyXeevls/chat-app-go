package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/zyXeevls/chat-app/internal/infrastructure/database"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file not found:", err)
	}

	db := database.NewPostgres()
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Realtime Chat Server Running.")
	})

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
