package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (h *UploadHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File error", 400)
		return
	}
	defer file.Close()

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		http.Error(w, "Could not read file", 500)
		return
	}

	contentType := http.DetectContentType(header[:n])
	var ext string

	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	default:
		ext = strings.ToLower(filepath.Ext(handler.Filename))
	}

	if ext == "" {
		ext = ".bin"
	}

	filename := uuid.New().String() + ext

	err = os.MkdirAll("uploads", 0o755)
	if err != nil {
		http.Error(w, "Could not prepare upload directory", 500)
		return
	}

	path := "uploads/" + filename

	dst, err := os.Create(path)
	if err != nil {
		http.Error(w, "Could not save file", 500)
		return
	}
	defer dst.Close()

	stream := io.MultiReader(bytes.NewReader(header[:n]), file)
	_, err = io.Copy(dst, stream)
	if err != nil {
		http.Error(w, "Could not save file", 500)
		return
	}

	resp := map[string]string{
		"filename": filename,
		"url":      os.Getenv("BASE_URL") + "/uploads/" + filename,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
