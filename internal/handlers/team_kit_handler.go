package handlers

import (
	"encoding/json"
	"net/http"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
	"strconv"

	"github.com/gorilla/mux"
)

// TeamKitHandler - handler dla zarządzania strojami zespołów
type TeamKitHandler struct {
	dbManager *database.Manager
}

// NewTeamKitHandler - tworzy nowy handler strojów
func NewTeamKitHandler(dbManager *database.Manager) *TeamKitHandler {
	return &TeamKitHandler{
		dbManager: dbManager,
	}
}

// GetTeamKits - pobiera stroje zespołu
// GET /api/teams/{id}/kits
func (h *TeamKitHandler) GetTeamKits(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Nieprawidłowe ID zespołu", http.StatusBadRequest)
		return
	}

	db := h.dbManager.GetDB()
	var kits []models.Kit

	if err := db.Preload("KitColors").Where("team_id = ?", teamID).Find(&kits).Error; err != nil {
		http.Error(w, "Błąd pobierania strojów", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"kits":   kits,
	})
}

// UpdateTeamKits - aktualizuje stroje zespołu
// PUT /api/teams/{id}/kits
func (h *TeamKitHandler) UpdateTeamKits(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Nieprawidłowe ID zespołu", http.StatusBadRequest)
		return
	}

	var kitsData map[string][]string
	if err := json.NewDecoder(r.Body).Decode(&kitsData); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	db := h.dbManager.GetDB()

	// Rozpocznij transakcję
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Pobierz istniejące stroje
	var kits []models.Kit
	if err := tx.Preload("KitColors").Where("team_id = ?", teamID).Find(&kits).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Błąd pobierania strojów", http.StatusInternalServerError)
		return
	}

	// Usuń wszystkie kolory dla każdego stroju
	for _, kit := range kits {
		if err := tx.Where("kit_id = ?", kit.ID).Delete(&models.KitColor{}).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Błąd usuwania kolorów", http.StatusInternalServerError)
			return
		}
	}

	// Aktualizuj kolory dla każdego typu stroju
	for kitTypeStr, colors := range kitsData {
		kitType, _ := strconv.Atoi(kitTypeStr)
		
		// Znajdź odpowiedni Kit
		var kit models.Kit
		for _, k := range kits {
			if k.Type == kitType {
				kit = k
				break
			}
		}

		// Dodaj nowe kolory
		for order, color := range colors {
			kitColor := models.KitColor{
				KitID:      kit.ID,
				ColorOrder: order + 1, // 1-based index
				Color:      color,
			}

			if err := tx.Create(&kitColor).Error; err != nil {
				tx.Rollback()
				http.Error(w, "Błąd dodawania koloru", http.StatusInternalServerError)
				return
			}
		}
	}

	// Zatwierdź transakcję
	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Błąd zatwierdzania transakcji", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Stroje zaktualizowane",
	})
}
