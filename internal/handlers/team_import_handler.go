package handlers

import (
	"encoding/json"
	"net/http"
	"recorder-server/internal/database"
	"recorder-server/internal/services"
	"recorder-server/internal/models"
//	"strconv"

	"github.com/gorilla/mux"
)

// TeamImportHandler - handler dla importu drużyn z pliku tymczasowego
type TeamImportHandler struct {
	importService *services.TeamImportService
	dbManager     *database.Manager
}

// NewTeamImportHandler - tworzy nowy handler importu
func NewTeamImportHandler(dbManager *database.Manager) *TeamImportHandler {
	return &TeamImportHandler{
		importService: services.NewTeamImportService(dbManager),
		dbManager:     dbManager,
	}
}

// GetTempTeams - pobiera listę tymczasowych drużyn
// GET /api/teams/temp
func (h *TeamImportHandler) GetTempTeams(w http.ResponseWriter, r *http.Request) {
	competitionID := h.dbManager.GetCurrentDatabaseName()
	
	teams, err := h.importService.GetTempTeams(competitionID)
	if err != nil {
		http.Error(w, "Błąd pobierania tymczasowych drużyn", http.StatusInternalServerError)
		return
	}

	stats, _ := h.importService.GetStatistics(competitionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"teams":      teams.Teams,
		"statistics": stats,
	})
}

// GetTempTeam - pobiera pojedynczą tymczasową drużynę
// GET /api/teams/temp/{temp_id}
func (h *TeamImportHandler) GetTempTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tempID := vars["temp_id"]
	
	competitionID := h.dbManager.GetCurrentDatabaseName()
	team, err := h.importService.GetTempTeam(competitionID, tempID)
	if err != nil {
		http.Error(w, "Nie znaleziono drużyny", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"team":   team,
	})
}

// UpdateTempTeam - aktualizuje tymczasową drużynę
// PUT /api/teams/temp/{temp_id}
func (h *TeamImportHandler) UpdateTempTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tempID := vars["temp_id"]

	var updated models.TempTeam
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	competitionID := h.dbManager.GetCurrentDatabaseName()
	if err := h.importService.UpdateTempTeam(competitionID, tempID, updated); err != nil {
		http.Error(w, "Błąd aktualizacji", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Drużyna zaktualizowana",
	})
}

// ImportTeam - importuje drużynę z pliku tymczasowego do bazy
// POST /api/teams/import/{temp_id}
func (h *TeamImportHandler) ImportTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tempID := vars["temp_id"]

	competitionID := h.dbManager.GetCurrentDatabaseName()
	team, err := h.importService.ImportTeam(competitionID, tempID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Drużyna zaimportowana",
		"team":    team,
	})
}

// ImportAllComplete - importuje wszystkie kompletne drużyny
// POST /api/teams/import-all
func (h *TeamImportHandler) ImportAllComplete(w http.ResponseWriter, r *http.Request) {
	competitionID := h.dbManager.GetCurrentDatabaseName()
	imported, errors := h.importService.ImportAllComplete(competitionID)

	response := map[string]interface{}{
		"status":   "success",
		"imported": imported,
	}

	if len(errors) > 0 {
		errorMessages := make([]string, len(errors))
		for i, err := range errors {
			errorMessages[i] = err.Error()
		}
		response["errors"] = errorMessages
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteTempTeam - usuwa tymczasową drużynę
// DELETE /api/teams/temp/{temp_id}
func (h *TeamImportHandler) DeleteTempTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tempID := vars["temp_id"]

	competitionID := h.dbManager.GetCurrentDatabaseName()
	if err := h.importService.DeleteTempTeam(competitionID, tempID); err != nil {
		http.Error(w, "Błąd usuwania", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Drużyna usunięta",
	})
}
