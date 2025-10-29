package scrapers

import (
	"fmt"
	"log"
	"sync"
)

// Registry - rejestr wszystkich dostępnych grup scraperów
type Registry struct {
	mu      sync.RWMutex
	groups  map[string]*ScraperGroup
}

var (
	registry *Registry
	once     sync.Once
)

// GetRegistry - singleton registry scraperów
func GetRegistry() *Registry {
	once.Do(func() {
		registry = &Registry{
			groups: make(map[string]*ScraperGroup),
		}
		// Automatyczna rejestracja wszystkich scraperów przy inicjalizacji
		registry.registerDefaultScrapers()
	})
	return registry
}

// RegisterGroup - rejestruje grupę scraperów
func (r *Registry) RegisterGroup(name string, group *ScraperGroup) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	group.Name = name
	r.groups[name] = group
	log.Printf("Scraper Registry: Zarejestrowano grupę scraperów '%s'", name)
}

// GetGroup - pobiera grupę scraperów po nazwie
func (r *Registry) GetGroup(name string) (*ScraperGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	group, exists := r.groups[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrScraperNotFound, name)
	}
	
	return group, nil
}

// ListGroups - zwraca listę dostępnych grup scraperów
func (r *Registry) ListGroups() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	groups := make([]string, 0, len(r.groups))
	for name := range r.groups {
		groups = append(groups, name)
	}
	return groups
}

// HasGroup - sprawdza czy grupa scraperów istnieje
func (r *Registry) HasGroup(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.groups[name]
	return exists
}

// registerDefaultScrapers - rejestruje domyślne scrapery
// W przyszłości tutaj będą rejestrowane konkretne implementacje
func (r *Registry) registerDefaultScrapers() {
	// Przykład rejestracji - w przyszłości będą tu prawdziwe implementacje
	
	// r.RegisterGroup("pzpn", &ScraperGroup{
	// 	TeamScraper:   &PZPNTeamScraper{},
	// 	PlayerScraper: &PZPNPlayerScraper{},
	// 	GameScraper:   &PZPNGameScraper{},
	// })
	
	// r.RegisterGroup("ekstraklasa", &ScraperGroup{
	// 	TeamScraper:   &EkstraklasaTeamScraper{},
	// 	PlayerScraper: &EkstraklasaPlayerScraper{},
	// 	GameScraper:   &EkstraklasaGameScraper{},
	// })
	
	log.Println("Scraper Registry: Inicjalizacja zakończona (brak domyślnych scraperów)")
}

// GetTeamScraper - pobiera scraper drużyn dla grupy
func (sg *ScraperGroup) GetTeamScraper() (TeamScraper, error) {
	if sg.TeamScraper == nil {
		return nil, ErrTeamScraperNotFound
	}
	return sg.TeamScraper, nil
}

// GetPlayerScraper - pobiera scraper zawodników dla grupy
func (sg *ScraperGroup) GetPlayerScraper() (PlayerScraper, error) {
	if sg.PlayerScraper == nil {
		return nil, ErrPlayerScraperNotFound
	}
	return sg.PlayerScraper, nil
}

// GetGameScraper - pobiera scraper meczów dla grupy
func (sg *ScraperGroup) GetGameScraper() (GameScraper, error) {
	if sg.GameScraper == nil {
		return nil, ErrGameScraperNotFound
	}
	return sg.GameScraper, nil
}
