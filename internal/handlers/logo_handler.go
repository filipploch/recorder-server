// internal/handlers/logo_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

type LogoHandler struct{}

func NewLogoHandler() *LogoHandler {
	return &LogoHandler{}
}

// ListLogos - zwraca listę dostępnych logo
func (h *LogoHandler) ListLogos(w http.ResponseWriter, r *http.Request) {
	logosPath := "./web/static/images/logos"

	files, err := os.ReadDir(logosPath)
	if err != nil {
		http.Error(w, "Błąd odczytu katalogu", http.StatusInternalServerError)
		return
	}

	var logos []map[string]string
	for _, file := range files {
		if !file.IsDir() && file.Name() != ".gitkeep" {
			ext := filepath.Ext(file.Name())
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".svg" {
				logos = append(logos, map[string]string{
					"name": file.Name(),
					"url":  "/static/images/logos/" + file.Name(),
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"logos":  logos,
	})
}
