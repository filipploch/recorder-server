package services

import (
	"encoding/json"
	"fmt"
	"log"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
	"recorder-server/internal/scrapers"
)

// ScraperService - serwis obsługujący operacje scrapowania
type ScraperService struct {
	dbManager *database.Manager
	registry  *scrapers.Registry
}

// NewScraperService - tworzy nowy serwis scraperów
func NewScraperService(dbManager *database.Manager) *ScraperService {
	return &ScraperService{
		dbManager: dbManager,
		registry:  scrapers.GetRegistry(),
	}
}

// getScraperNameFromVariable - pobiera nazwę scrapera z pola Variable
func (s *ScraperService) getScraperNameFromVariable(variable string) (string, error) {
	var variableData map[string]interface{}
	if err := json.Unmarshal([]byte(variable), &variableData); err != nil {
		return "", fmt.Errorf("błąd parsowania Variable: %w", err)
	}
	
	// Sprawdź czy istnieje pole scraper.name
	if scraper, ok := variableData["scraper"].(map[string]interface{}); ok {
		if name, ok := scraper["name"].(string); ok && name != "" {
			return name, nil
		}
	}
	
	return "", fmt.Errorf("brak informacji o scraperze w Variable")
}

// ScrapeTeamsForCompetition - scrapuje drużyny dla danej konkurencji
func (s *ScraperService) ScrapeTeamsForCompetition(competitionID uint, externalCompetitionID string) ([]models.Team, error) {
	// Pobierz competition z bazy
	db := s.dbManager.GetDB()
	var competition models.Competition
	if err := db.First(&competition, competitionID).Error; err != nil {
		return nil, fmt.Errorf("nie znaleziono competition: %w", err)
	}
	
	// Pobierz nazwę scrapera z Variable
	scraperName, err := s.getScraperNameFromVariable(competition.Variable)
	if err != nil {
		return nil, fmt.Errorf("brak przypisanego scrapera dla competition ID=%d: %w", competitionID, err)
	}
	
	log.Printf("ScraperService: Pobieranie drużyn dla competition '%s' używając scrapera '%s'",
		competition.Name, scraperName)
	
	// Pobierz grupę scraperów
	group, err := s.registry.GetGroup(scraperName)
	if err != nil {
		return nil, fmt.Errorf("błąd pobierania scrapera: %w", err)
	}
	
	// Pobierz scraper drużyn
	teamScraper, err := group.GetTeamScraper()
	if err != nil {
		return nil, fmt.Errorf("brak scrapera drużyn: %w", err)
	}
	
	// Wykonaj scrapowanie
	log.Printf("ScraperService: Uruchamianie scrapera '%s'...", teamScraper.GetName())
	teams, err := teamScraper.ScrapeTeams(externalCompetitionID)
	if err != nil {
		return nil, fmt.Errorf("błąd scrapowania: %w", err)
	}
	
	log.Printf("ScraperService: Pobrano %d drużyn", len(teams))
	return teams, nil
}

// ScrapePlayersForTeam - scrapuje zawodników dla danej drużyny
func (s *ScraperService) ScrapePlayersForTeam(teamID uint, externalTeamID string) ([]models.Player, error) {
	// Pobierz drużynę z bazy
	db := s.dbManager.GetDB()
	var team models.Team
	if err := db.First(&team, teamID).Error; err != nil {
		return nil, fmt.Errorf("nie znaleziono team: %w", err)
	}
	
	// Sprawdź czy drużyna ma ForeignID (dla śledzenia źródła)
	if team.ForeignID == nil {
		log.Printf("ScraperService: Uwaga - team ID=%d nie ma ForeignID, używam parametru externalTeamID", teamID)
	}
	
	// Znajdź competition dla tej drużyny (przez GameTeam -> Game -> Group -> Stage -> Competition)
	var competition models.Competition
	if err := db.Raw(`
		SELECT DISTINCT c.* 
		FROM competitions c
		JOIN stages s ON s.competition_id = c.id
		JOIN groups g ON g.stage_id = s.id
		JOIN group_teams gt ON gt.group_id = g.id
		WHERE gt.team_id = ?
		LIMIT 1
	`, teamID).Scan(&competition).Error; err != nil {
		return nil, fmt.Errorf("nie znaleziono competition dla team: %w", err)
	}
	
	// Pobierz nazwę scrapera z Variable
	scraperName, err := s.getScraperNameFromVariable(competition.Variable)
	if err != nil {
		return nil, fmt.Errorf("brak przypisanego scrapera dla competition związanej z team ID=%d: %w", teamID, err)
	}
	
	log.Printf("ScraperService: Pobieranie zawodników dla team '%s' używając scrapera '%s'",
		team.Name, scraperName)
	
	// Pobierz grupę scraperów
	group, err := s.registry.GetGroup(scraperName)
	if err != nil {
		return nil, fmt.Errorf("błąd pobierania scrapera: %w", err)
	}
	
	// Pobierz scraper zawodników
	playerScraper, err := group.GetPlayerScraper()
	if err != nil {
		return nil, fmt.Errorf("brak scrapera zawodników: %w", err)
	}
	
	// Wykonaj scrapowanie
	log.Printf("ScraperService: Uruchamianie scrapera '%s'...", playerScraper.GetName())
	players, err := playerScraper.ScrapePlayers(teamID, externalTeamID)
	if err != nil {
		return nil, fmt.Errorf("błąd scrapowania: %w", err)
	}
	
	log.Printf("ScraperService: Pobrano %d zawodników", len(players))
	return players, nil
}

// ScrapeGamesForStage - scrapuje mecze dla danego stage
func (s *ScraperService) ScrapeGamesForStage(stageID uint, externalCompetitionID string) ([]models.Game, error) {
	// Pobierz stage z bazy
	db := s.dbManager.GetDB()
	var stage models.Stage
	if err := db.Preload("Competition").First(&stage, stageID).Error; err != nil {
		return nil, fmt.Errorf("nie znaleziono stage: %w", err)
	}
	
	competition := stage.Competition
	
	// Pobierz nazwę scrapera z Variable
	scraperName, err := s.getScraperNameFromVariable(competition.Variable)
	if err != nil {
		return nil, fmt.Errorf("brak przypisanego scrapera dla stage ID=%d: %w", stageID, err)
	}
	
	log.Printf("ScraperService: Pobieranie meczów dla stage '%s' używając scrapera '%s'",
		stage.Name, scraperName)
	
	// Pobierz grupę scraperów
	group, err := s.registry.GetGroup(scraperName)
	if err != nil {
		return nil, fmt.Errorf("błąd pobierania scrapera: %w", err)
	}
	
	// Pobierz scraper meczów
	gameScraper, err := group.GetGameScraper()
	if err != nil {
		return nil, fmt.Errorf("brak scrapera meczów: %w", err)
	}
	
	// Wykonaj scrapowanie
	log.Printf("ScraperService: Uruchamianie scrapera '%s'...", gameScraper.GetName())
	games, err := gameScraper.ScrapeGames(externalCompetitionID, stageID)
	if err != nil {
		return nil, fmt.Errorf("błąd scrapowania: %w", err)
	}
	
	log.Printf("ScraperService: Pobrano %d meczów", len(games))
	return games, nil
}

// GetAvailableScrapers - zwraca listę dostępnych scraperów
func (s *ScraperService) GetAvailableScrapers() []string {
	return s.registry.ListGroups()
}