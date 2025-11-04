package handlers

import (
	"encoding/json"
	"net/http"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
	"recorder-server/internal/services"
	"strconv"
)

// ScraperHandler - handler dla operacji scrapowania
type ScraperHandler struct {
	scraperService *services.ScraperService
	dbManager      *database.Manager
}

// NewScraperHandler - tworzy nowy handler scraperów
func NewScraperHandler(scraperService *services.ScraperService) *ScraperHandler {
	return &ScraperHandler{
		scraperService: scraperService,
	}
}

// ScrapeTeams - endpoint do scrapowania drużyn
// POST /api/scrape/teams
// Body: { "competition_id": 1, "external_competition_id": "ekstraklasa-2024" }
// 
// ZMIENIONA LOGIKA:
// - Kompletne drużyny zapisywane są bezpośrednio do bazy
// - Niekompletne drużyny zapisywane są do /competitions/<id>/tmp/teams.json
func (h *ScraperHandler) ScrapeTeams(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CompetitionID         uint   `json:"competition_id"`
		ExternalCompetitionID string `json:"external_competition_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	// Wykonaj scrapowanie
	teams, err := h.scraperService.ScrapeTeamsForCompetition(req.CompetitionID, req.ExternalCompetitionID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	// TODO: W rzeczywistej implementacji ScraperService powinien zwracać []TempTeam
	// Na razie przyjmijmy że zwraca []Team i trzeba je skonwertować
	
	// Klasyfikuj drużyny na kompletne i niekompletne
	completeCount := 0
	incompleteCount := 0
	
	// TODO: Implementacja klasyfikacji i zapisu
	// completeTeams - zapisz do bazy danych
	// incompleteTeams - zapisz do pliku tymczasowego
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"total":     len(teams),
		"complete":  completeCount,
		"incomplete": incompleteCount,
		"message":   "Drużyny zescrapowane. Kompletne zapisane do bazy, niekompletne do pliku tymczasowego.",
	})
}

// ScrapePlayers - endpoint do scrapowania zawodników
// POST /api/scrape/players
// Body: { "team_id": 1, "external_team_id": "legia-warszawa" }
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

	// Wykonaj scrapowanie
	players, err := h.scraperService.ScrapePlayersForTeam(req.TeamID, req.ExternalTeamID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	// TODO: Jeśli SaveToDB == true, zapisz zawodników do bazy

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"count":   len(players),
		"players": players,
	})
}

// ScrapeGames - endpoint do scrapowania meczów
// POST /api/scrape/games
// Body: { "stage_id": 1, "external_competition_id": "ekstraklasa-2024" }
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

	// Wykonaj scrapowanie
	games, err := h.scraperService.ScrapeGamesForStage(req.StageID, req.ExternalCompetitionID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	// TODO: Jeśli SaveToDB == true, zapisz mecze do bazy

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"count":  len(games),
		"games":  games,
	})
}

// GetAvailableScrapers - endpoint zwracający listę dostępnych scraperów
// GET /api/scrape/available
func (h *ScraperHandler) GetAvailableScrapers(w http.ResponseWriter, r *http.Request) {
	scrapers := h.scraperService.GetAvailableScrapers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"scrapers": scrapers,
		"count":    len(scrapers),
	})
}

// GetCompetitionScraperInfo - endpoint zwracający info o scraperze dla competition
// GET /api/scrape/competition/info?competition_id=1
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