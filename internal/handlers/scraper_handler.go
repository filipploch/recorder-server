package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
	"recorder-server/internal/scrapers"
	"recorder-server/internal/services"
	"strconv"
)

// ScraperHandler - handler dla operacji scrapowania
type ScraperHandler struct {
	scraperService *services.ScraperService
	dbManager      *database.Manager // Dodaj to pole
}

// NewScraperHandler - tworzy nowy handler scraperów
func NewScraperHandler(dbManager *database.Manager) *ScraperHandler {
	return &ScraperHandler{
		dbManager: dbManager, // Ustaw dbManager
	}
}

// ScrapeTeams - endpoint do scrapowania drużyn
func (h *ScraperHandler) ScrapeTeams(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CompetitionID uint `json:"competition_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	db := h.dbManager.GetDB()
	if db == nil {
		http.Error(w, "Brak połączenia z bazą danych", http.StatusInternalServerError)
		return
	}

	var competition models.Competition
	if err := db.First(&competition, req.CompetitionID).Error; err != nil {
		http.Error(w, "Nie znaleziono competition", http.StatusNotFound)
		return
	}

	var variableData map[string]interface{}
	if err := json.Unmarshal([]byte(competition.Variable), &variableData); err != nil {
		http.Error(w, "Błąd parsowania Variable", http.StatusBadRequest)
		return
	}

	scraperData, ok := variableData["scraper"].(map[string]interface{})
	if !ok {
		http.Error(w, "Brak danych scrapera w Variable", http.StatusBadRequest)
		return
	}

	scraperName, ok := scraperData["name"].(string)
	if !ok || scraperName == "" {
		http.Error(w, "Brak nazwy scrapera", http.StatusBadRequest)
		return
	}

	teamsURL, ok := scraperData["teams_url"].(string)
	if !ok || teamsURL == "" {
		http.Error(w, "Brak teams_url w scraperze", http.StatusBadRequest)
		return
	}

	// Pobierz scraper
	registry := scrapers.GetRegistry()
	group, err := registry.GetGroup(scraperName)
	if err != nil {
		http.Error(w, "Nie znaleziono scrapera: "+scraperName, http.StatusNotFound)
		return
	}

	// Sprawdź czy scraper implementuje TeamScraperWithDB
	teamScraperWithDB, ok := group.TeamScraper.(scrapers.TeamScraperWithDB)
	if !ok {
		http.Error(w, "Scraper nie obsługuje scrapowania z bazą danych", http.StatusInternalServerError)
		return
	}

	// Wykonaj scrapowanie z przekazaniem bazy danych
	competitionID := h.dbManager.GetCurrentDatabaseName()
	tempTeams, err := teamScraperWithDB.ScrapeTeamsWithDB(competitionID, teamsURL, db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	// Zapisz do pliku tymczasowego
	manager := models.NewTempTeamManager(competitionID)
	if err := manager.AddBulk(tempTeams); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  "Błąd zapisu: " + err.Error(),
		})
		return
	}

	// Pobierz statystyki
	collection, _ := manager.Load()
	complete := 0
	incomplete := 0
	for _, team := range collection.Teams {
		if team.IsComplete() {
			complete++
		} else {
			incomplete++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"total":      len(collection.Teams),
		"complete":   complete,
		"incomplete": incomplete,
		"new":        len(tempTeams),
		"message":    fmt.Sprintf("Dodano %d nowych drużyn do pliku tymczasowego", len(tempTeams)),
	})
}

// ScrapePlayers - endpoint do scrapowania zawodników
func (h *ScraperHandler) ScrapePlayers(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeamID         uint   `json:"team_id"`
		ExternalTeamID string `json:"external_team_id"`
		SaveToDB       bool   `json:"save_to_db"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implementacja scrapowania zawodników

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "error",
		"message": "Not implemented yet",
	})
}

// ScrapeGames - endpoint do scrapowania meczów
func (h *ScraperHandler) ScrapeGames(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StageID               uint   `json:"stage_id"`
		ExternalCompetitionID string `json:"external_competition_id"`
		SaveToDB              bool   `json:"save_to_db"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	// TODO: Implementacja scrapowania meczów

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "error",
		"message": "Not implemented yet",
	})
}

// GetAvailableScrapers - endpoint zwracający listę dostępnych scraperów
func (h *ScraperHandler) GetAvailableScrapers(w http.ResponseWriter, r *http.Request) {
	registry := scrapers.GetRegistry()
	scrapersList := registry.ListGroups()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"scrapers": scrapersList,
		"count":    len(scrapersList),
	})
}

// GetCompetitionScraperInfo - endpoint zwracający info o scraperze dla competition
func (h *ScraperHandler) GetCompetitionScraperInfo(w http.ResponseWriter, r *http.Request) {
	competitionIDStr := r.URL.Query().Get("competition_id")
	if competitionIDStr == "" {
		http.Error(w, "Brak competition_id", http.StatusBadRequest)
		return
	}

	competitionID, err := strconv.ParseUint(competitionIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Nieprawidłowe competition_id", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "success",
		"competition_id": competitionID,
		"message":        "Informacje o scraperze są przechowywane w polu Variable",
	})
}
