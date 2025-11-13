package scrapers

import (
	"fmt"
	"recorder-server/internal/models"

	"gorm.io/gorm"
)

// ExampleScraper - przykładowa implementacja scrapera (placeholder)
type ExampleScraper struct {
	name string
}

// NewExampleScraper - tworzy nowy przykładowy scraper
func NewExampleScraper(name string) *ExampleScraper {
	return &ExampleScraper{
		name: name,
	}
}

// GetName - implementacja interfejsu Scraper
func (s *ExampleScraper) GetName() string {
	return s.name
}

// ScrapeTeams - implementacja interfejsu TeamScraper
func (s *ExampleScraper) ScrapeTeams(competitionID string) ([]models.Team, error) {
	return nil, fmt.Errorf("not implemented: %s team scraper", s.name)
}

// ScrapePlayers - implementacja interfejsu PlayerScraper
func (s *ExampleScraper) ScrapePlayers(teamID uint, externalTeamID string) ([]models.Player, error) {
	return nil, fmt.Errorf("not implemented: %s player scraper", s.name)
}

// ScrapeGames - implementacja interfejsu GameScraper
func (s *ExampleScraper) ScrapeGames(competitionID string, stageID uint) ([]models.Game, error) {
	return nil, fmt.Errorf("not implemented: %s game scraper", s.name)
}

// MZPNScraper jest teraz zdefiniowany w mzpn_team_scraper.go jako MZPNScraperWrapper
// Ta funkcja tworzy nową instancję wrappera
func NewMZPNScraper() *MZPNScraperWrapper {
	return NewMZPNScraperWrapper()
}

// NalffutsalScraper - scraper dla rozgrywek NALF
type NalffutsalScraper struct {
	*ExampleScraper
	baseURL     string
	teamScraper *NalffutsalTeamScraper
}

// NewNalffutsalScraper - tworzy scraper dla NALF
func NewNalffutsalScraper() *NalffutsalScraper {
	return &NalffutsalScraper{
		ExampleScraper: NewExampleScraper("Nalffutsal"),
		baseURL:        "https://nalffutsal.pl",
		teamScraper:    NewNalffutsalTeamScraper(),
	}
}

// ScrapeTeams - implementacja dla NALF (stara sygnatura - nieużywana)
func (s *NalffutsalScraper) ScrapeTeams(competitionID string) ([]models.Team, error) {
	return nil, fmt.Errorf("use ScrapeTeamsWithDB instead")
}

// ScrapeTeamsWithDB - nowa metoda używana przez handler
func (s *NalffutsalScraper) ScrapeTeamsWithDB(competitionID string, teamsURL string, db *gorm.DB) ([]models.TempTeam, error) {
	// Użyj dedykowanego team scrapera
	tempTeams, err := s.teamScraper.ScrapeTeams(competitionID, teamsURL, db)
	if err != nil {
		return nil, fmt.Errorf("NALF team scraper error: %w", err)
	}

	// Wyświetl podsumowanie
	s.teamScraper.PrintSummary(tempTeams)

	return tempTeams, nil
}

// RegisterExampleScrapers - rejestruje przykładowe scrapery
func RegisterExampleScrapers() {
	registry := GetRegistry()

	// Rejestracja scrapera MZPN
	mzpnScraper := NewMZPNScraper()
	registry.RegisterGroup("mzpn", &ScraperGroup{
		TeamScraper:   mzpnScraper,
		PlayerScraper: mzpnScraper,
		GameScraper:   mzpnScraper,
	})

	// Rejestracja scrapera Nalffutsal
	nalffutsalScraper := NewNalffutsalScraper()
	registry.RegisterGroup("nalffutsal", &ScraperGroup{
		TeamScraper:   nalffutsalScraper,
		PlayerScraper: nalffutsalScraper,
		GameScraper:   nalffutsalScraper,
	})
}
