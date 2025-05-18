package webfront

import (
	"encoding/json"
	"fmt"
	"github.com/paul-at-nangalan/errorhandler/handlers"
	"io"
	"net/http"
	"object-detection-zero-shot/middleware"
	"object-detection-zero-shot/service"
	"os"
	"path/filepath"
	"strings"
)

type Handler struct {
	svc       *service.Handler
	uploadDir string
}

func NewHandler(svc *service.Handler, uploadDir string) *Handler {
	h := &Handler{
		svc:       svc,
		uploadDir: uploadDir,
	}

	throttleEmbed := middleware.NewThrottleMiddleware(30, 24)
	http.HandleFunc("/image/embed", throttleEmbed.Wrap(h.HandleImageUpload))
	throttleDetect := middleware.NewThrottleMiddleware(30, 24)
	http.HandleFunc("/image/detect", throttleDetect.Wrap(h.HandleImageDetection))

	http.Handle("/", http.FileServer(http.Dir("/webfront/static")))
	return h
}

func (h *Handler) HandleImageUpload(w http.ResponseWriter, r *http.Request) {
	defer handlers.NetHandlePanic(w)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse multipart form with 10MB max memory
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	// Get the file from form data
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()
	// Get associated text data
	text := r.FormValue("text")
	if text == "" {
		http.Error(w, "Text description is required", http.StatusBadRequest)
		return
	}
	// Create sanitized ID from filename
	filename := header.Filename
	ext := filepath.Ext(filename)
	baseFilename := strings.TrimSuffix(filename, ext)
	sanitizedID := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, baseFilename)
	// Save file to disk
	newFilename := sanitizedID + ext
	filepath := filepath.Join(h.uploadDir, newFilename)
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err = io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	// Create embedding configuration
	embedCfg := &service.EmbedCfg{
		Items: []service.Item{
			{
				Imagefile: filepath,
				Label:     text,
				ID:        sanitizedID,
			},
		},
	}
	// Create embeddings
	h.svc.EmbedData(embedCfg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	resp := map[string]any{
		"status": "success",
		"id":     sanitizedID,
	}
	err = enc.Encode(resp)
	if err != nil {
		fmt.Printf("Failed to send response with error %s", err.Error())
	}
}

type DectionResponse struct {
	Found bool    `json:"found"`
	Label string  `json:"label"`
	Score float32 `json:"score"`
}

func (h *Handler) HandleImageDetection(w http.ResponseWriter, r *http.Request) {
	defer handlers.NetHandlePanic(w)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse multipart form with 10MB max memory
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	// Get the file from form data
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()
	// Create sanitized ID from filename
	filename := header.Filename
	ext := filepath.Ext(filename)
	baseFilename := strings.TrimSuffix(filename, ext)
	sanitizedID := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, baseFilename)
	// Save file to disk
	newFilename := sanitizedID + ext
	filepath := filepath.Join(h.uploadDir, newFilename)
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err = io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	// Perform image detection
	results := h.svc.ImageDetection(filepath)

	resp := DectionResponse{
		Found: false,
	}
	if len(results) > 0 {
		resp.Found = true
		resp.Label = results[0].Metadata["value"].(string)
		resp.Score = results[0].Score
	}

	// Return results as JSON
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		fmt.Println("Error writing response ", err)
	}
}
