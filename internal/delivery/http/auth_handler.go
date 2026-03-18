package http

import (
	"encoding/json"
	"net/http"

	"github.com/zyXeevls/chat-app/internal/repository"
	"github.com/zyXeevls/chat-app/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo *repository.AuthRepository
}

func NewAuthHandler(repo *repository.AuthRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&body)

	err := h.repo.Register(body.Username, body.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte("register success"))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&body)

	id, hash, err := h.repo.GetUser(body.Username)
	if err != nil {
		http.Error(w, "User not found", 404)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password))
	if err != nil {
		http.Error(w, "wrong password", 401)
		return
	}

	token, _ := utils.GenerateToken(id)

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
