package scrapers

import (
	"errors"
	"recorder-server/internal/models"
)

// Scraper - interfejs dla wszystkich scraperów
type Scraper interface {
	GetName() string // Nazwa scrapera dla logowania
}

// TeamScraper - interfejs dla scraperów pobierających drużyny
type TeamScraper interface {
	Scraper
	ScrapeTeams(competitionID string) ([]models.Team, error)
}

// PlayerScraper - interfejs dla scraperów pobierających zawodników
type PlayerScraper interface {
	Scraper
	ScrapePlayers(teamID uint, externalTeamID string) ([]models.Player, error)
}

// GameScraper - interfejs dla scraperów pobierających mecze
type GameScraper interface {
	Scraper
	ScrapeGames(competitionID string, stageID uint) ([]models.Game, error)
}

// FullScraper - interfejs dla scraperów obsługujących wszystkie operacje
type FullScraper interface {
	TeamScraper
	PlayerScraper
	GameScraper
}

// ScraperGroup - grupa scraperów dla konkretnych rozgrywek
type ScraperGroup struct {
	Name          string
	TeamScraper   TeamScraper
	PlayerScraper PlayerScraper
	GameScraper   GameScraper
}

// Errors
var (
	ErrScraperNotFound      = errors.New("scraper group not found")
	ErrTeamScraperNotFound  = errors.New("team scraper not available for this group")
	ErrPlayerScraperNotFound = errors.New("player scraper not available for this group")
	ErrGameScraperNotFound  = errors.New("game scraper not available for this group")
	ErrScrapingFailed       = errors.New("scraping operation failed")
)
