package handlers

import (
	"encoding/json"
	"net/http"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
	"recorder-server/internal/services"

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

	collection, err := h.importService.GetTempTeams(competitionID)
	if err != nil {
		http.Error(w, "Błąd pobierania tymczasowych drużyn", http.StatusInternalServerError)
		return
	}

	stats, _ := h.importService.GetStatistics(competitionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"teams":      collection.Teams, // ZMIENIONE - zwróć collection.Teams zamiast collection
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

	// Użyj struktury pomocniczej do odbierania danych z frontendu
	var updateData struct {
		Name      string              `json:"name"`
		ShortName string              `json:"short_name"`
		Name16    string              `json:"name_16"`
		Logo      string              `json:"logo"`
		Link      string              `json:"link"`
		ForeignID string              `json:"foreign_id"`
		Source    string              `json:"source"`
		ScrapedAt string              `json:"scraped_at"`
		Notes     string              `json:"notes"`
		Kits      map[string][]string `json:"kits"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Błąd dekodowania JSON: " + err.Error(),
		})
		return
	}

	competitionID := h.dbManager.GetCurrentDatabaseName()

	// Pobierz istniejącą drużynę
	existing, err := h.importService.GetTempTeam(competitionID, tempID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Nie znaleziono drużyny: " + err.Error(),
		})
		return
	}

	// Konwertuj na TempTeam ze wskaźnikami
	updated := models.TempTeam{
		TempID:    existing.TempID,
		Name:      updateData.Name,
		ShortName: stringToPtr(updateData.ShortName),
		Name16:    stringToPtr(updateData.Name16),
		Logo:      stringToPtr(updateData.Logo),
		Link:      stringToPtr(updateData.Link),
		ForeignID: stringToPtr(updateData.ForeignID),
		Source:    existing.Source, // Zachowaj oryginalne źródło
		ScrapedAt: existing.ScrapedAt,
		Notes:     updateData.Notes,
		Kits:      updateData.Kits,
	}

	// Sprawdź czy drużyna jest kompletna
	if updated.IsComplete() {
		// Najpierw zapisz w pliku tymczasowym
		if err := h.importService.UpdateTempTeam(competitionID, tempID, updated); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "error",
				"error":  "Błąd zapisu: " + err.Error(),
			})
			return
		}

		// Następnie importuj do bazy
		team, err := h.importService.ImportTeam(competitionID, tempID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "error",
				"error":  "Błąd importu: " + err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "success",
			"message":  "Drużyna zaimportowana do bazy danych",
			"team":     team,
			"imported": true,
		})
		return
	}

	// Drużyna niekompletna - zapisz w pliku tymczasowym
	if err := h.importService.UpdateTempTeam(competitionID, tempID, updated); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Błąd aktualizacji: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"message":  "Drużyna zaktualizowana w pliku tymczasowym",
		"imported": false,
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
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
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

// stringToPtr - helper do konwersji string na *string (nil jeśli pusty)
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
