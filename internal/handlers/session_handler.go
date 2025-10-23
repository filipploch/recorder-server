package handlers

import (
	"encoding/json"
	"net/http"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
)

// SessionHandler - handler dla aktywnej sesji
type SessionHandler struct {
	manager *database.Manager
}

// NewSessionHandler - tworzy nowy handler sesji
func NewSessionHandler(manager *database.Manager) *SessionHandler {
	return &SessionHandler{
		manager: manager,
	}
}

// GetActiveSession - pobiera aktywną sesję
func (h *SessionHandler) GetActiveSession(w http.ResponseWriter, r *http.Request) {
	db := h.manager.GetDB()
	
	var session models.ActiveSession
	
	// Pobierz pierwszy (i jedyny) rekord
	if err := db.Preload("Game").Preload("GamePart").First(&session).Error; err != nil {
		// Jeśli nie ma rekordu, utwórz pusty
		session = models.ActiveSession{
			GameID:     nil,
			GamePartID: nil,
		}
		db.Create(&session)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"session": session,
	})
}

// SetActiveGame - ustawia aktywny mecz
func (h *SessionHandler) SetActiveGame(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GameID *uint `json:"game_id"` // nullable
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	db := h.manager.GetDB()
	
	var session models.ActiveSession
	
	// Pobierz lub utwórz sesję
	if err := db.First(&session).Error; err != nil {
		session = models.ActiveSession{}
		db.Create(&session)
	}
	
	// Aktualizuj GameID
	session.GameID = req.GameID
	session.GamePartID = nil // Resetuj część meczu przy zmianie meczu
	
	if err := db.Save(&session).Error; err != nil {
		http.Error(w, "Błąd zapisywania sesji", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Aktywny mecz ustawiony",
		"session": session,
	})
}

// SetActiveGamePart - ustawia aktywną część meczu
func (h *SessionHandler) SetActiveGamePart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GamePartID *uint `json:"game_part_id"` // nullable
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	db := h.manager.GetDB()
	
	var session models.ActiveSession
	
	// Pobierz lub utwórz sesję
	if err := db.First(&session).Error; err != nil {
		session = models.ActiveSession{}
		db.Create(&session)
	}
	
	// Aktualizuj GamePartID
	session.GamePartID = req.GamePartID
	
	if err := db.Save(&session).Error; err != nil {
		http.Error(w, "Błąd zapisywania sesji", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Aktywna część meczu ustawiona",
		"session": session,
	})
}

// ClearActiveSession - czyści aktywną sesję
func (h *SessionHandler) ClearActiveSession(w http.ResponseWriter, r *http.Request) {
	db := h.manager.GetDB()
	
	var session models.ActiveSession
	
	// Pobierz sesję
	if err := db.First(&session).Error; err != nil {
		http.Error(w, "Nie znaleziono sesji", http.StatusNotFound)
		return
	}
	
	// Wyczyść pola
	session.GameID = nil
	session.GamePartID = nil
	
	if err := db.Save(&session).Error; err != nil {
		http.Error(w, "Błąd zapisywania sesji", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Sesja wyczyszczona",
		"session": session,
	})
}
