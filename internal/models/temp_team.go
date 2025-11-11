package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TempTeam - tymczasowa drużyna ze scrapowanych danych
type TempTeam struct {
	TempID      string  `json:"temp_id"`       // Unikalny ID tymczasowe (UUID)
	Name        string  `json:"name"`          // Pełna nazwa (wymagana)
	ShortName   *string `json:"short_name"`    // Skrót 3-znakowy (opcjonalny)
	Name16      *string `json:"name_16"`       // Nazwa 16-znakowa (opcjonalny)
	Logo        *string `json:"logo"`          // URL/ścieżka do logo (opcjonalny)
	Link        *string `json:"link"`          // Link źródłowy (opcjonalny)
	ForeignID   *string `json:"foreign_id"`    // ID zewnętrzne
	Source      string  `json:"source"`        // Źródło danych (np. "mzpn_scraper")
	ScrapedAt   string  `json:"scraped_at"`    // Data scrapowania
	Notes       string  `json:"notes"`         // Dodatkowe notatki
}

// TempTeamsCollection - kolekcja tymczasowych drużyn
type TempTeamsCollection struct {
	Teams []TempTeam `json:"teams"`
}

// TempTeamManager - zarządza plikiem tymczasowych drużyn
type TempTeamManager struct {
	competitionID string
	filePath      string
	mu            sync.RWMutex
}

// NewTempTeamManager - tworzy nowy manager dla danej konkurencji
func NewTempTeamManager(competitionID string) *TempTeamManager {
	basePath := filepath.Join("./competitions", competitionID, "tmp")
	os.MkdirAll(basePath, 0755)
	
	return &TempTeamManager{
		competitionID: competitionID,
		filePath:      filepath.Join(basePath, "teams.json"),
	}
}

// Load - wczytuje tymczasowe drużyny z pliku
func (m *TempTeamManager) Load() (*TempTeamsCollection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Sprawdź czy plik istnieje
	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		// Zwróć pustą kolekcję
		return &TempTeamsCollection{Teams: []TempTeam{}}, nil
	}
	
	data, err := ioutil.ReadFile(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("błąd odczytu pliku: %w", err)
	}
	
	var collection TempTeamsCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("błąd parsowania JSON: %w", err)
	}
	
	return &collection, nil
}

// Save - zapisuje tymczasowe drużyny do pliku
func (m *TempTeamManager) Save(collection *TempTeamsCollection) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("błąd serializacji JSON: %w", err)
	}
	
	if err := ioutil.WriteFile(m.filePath, data, 0644); err != nil {
		return fmt.Errorf("błąd zapisu pliku: %w", err)
	}
	
	return nil
}

// Add - dodaje nową tymczasową drużynę
func (m *TempTeamManager) Add(team TempTeam) error {
	collection, err := m.Load()
	if err != nil {
		return err
	}
	
	// Sprawdź czy już istnieje
	for _, t := range collection.Teams {
		if t.TempID == team.TempID {
			return fmt.Errorf("drużyna o ID %s już istnieje w pliku tymczasowym", team.TempID)
		}
	}
	
	// Ustaw czas scrapowania jeśli nie podano
	if team.ScrapedAt == "" {
		team.ScrapedAt = time.Now().Format(time.RFC3339)
	}
	
	collection.Teams = append(collection.Teams, team)
	return m.Save(collection)
}

// AddBulk - dodaje wiele drużyn na raz
func (m *TempTeamManager) AddBulk(teams []TempTeam) error {
	collection, err := m.Load()
	if err != nil {
		return err
	}
	
	// Mapa istniejących ID
	existing := make(map[string]bool)
	for _, t := range collection.Teams {
		existing[t.TempID] = true
	}
	
	// Dodaj tylko nowe
	added := 0
	for _, team := range teams {
		if !existing[team.TempID] {
			if team.ScrapedAt == "" {
				team.ScrapedAt = time.Now().Format(time.RFC3339)
			}
			collection.Teams = append(collection.Teams, team)
			added++
		}
	}
	
	if added > 0 {
		return m.Save(collection)
	}
	
	return nil
}

// Get - pobiera drużynę po TempID
func (m *TempTeamManager) Get(tempID string) (*TempTeam, error) {
	collection, err := m.Load()
	if err != nil {
		return nil, err
	}
	
	for _, team := range collection.Teams {
		if team.TempID == tempID {
			return &team, nil
		}
	}
	
	return nil, fmt.Errorf("nie znaleziono drużyny o ID: %s", tempID)
}

// Update - aktualizuje drużynę tymczasową
func (m *TempTeamManager) Update(tempID string, updated TempTeam) error {
	collection, err := m.Load()
	if err != nil {
		return err
	}
	
	found := false
	for i, team := range collection.Teams {
		if team.TempID == tempID {
			updated.TempID = tempID // Zachowaj oryginalne ID
			updated.ScrapedAt = team.ScrapedAt // Zachowaj czas scrapowania
			collection.Teams[i] = updated
			found = true
			break
		}
	}
	
	if !found {
		return fmt.Errorf("nie znaleziono drużyny o ID: %s", tempID)
	}
	
	return m.Save(collection)
}

// Delete - usuwa drużynę po TempID
func (m *TempTeamManager) Delete(tempID string) error {
	collection, err := m.Load()
	if err != nil {
		return err
	}
	
	newTeams := []TempTeam{}
	found := false
	
	for _, team := range collection.Teams {
		if team.TempID != tempID {
			newTeams = append(newTeams, team)
		} else {
			found = true
		}
	}
	
	if !found {
		return fmt.Errorf("nie znaleziono drużyny o ID: %s", tempID)
	}
	
	collection.Teams = newTeams
	return m.Save(collection)
}

// Count - zwraca liczbę tymczasowych drużyn
func (m *TempTeamManager) Count() (int, error) {
	collection, err := m.Load()
	if err != nil {
		return 0, err
	}
	return len(collection.Teams), nil
}

// Clear - usuwa wszystkie tymczasowe drużyny
func (m *TempTeamManager) Clear() error {
	collection := &TempTeamsCollection{Teams: []TempTeam{}}
	return m.Save(collection)
}

// IsComplete - sprawdza czy drużyna ma wszystkie wymagane dane
func (t *TempTeam) IsComplete() bool {
	return t.Name != "" &&
		t.ShortName != nil && *t.ShortName != "" &&
		t.Name16 != nil && *t.Name16 != "" &&
		t.Logo != nil && *t.Logo != ""
}

// GetMissingFields - zwraca listę brakujących pól
func (t *TempTeam) GetMissingFields() []string {
	missing := []string{}
	
	if t.Name == "" {
		missing = append(missing, "name")
	}
	if t.ShortName == nil || *t.ShortName == "" {
		missing = append(missing, "short_name")
	}
	if t.Name16 == nil || *t.Name16 == "" {
		missing = append(missing, "name_16")
	}
	if t.Logo == nil || *t.Logo == "" {
		missing = append(missing, "logo")
	}
	
	return missing
}

// ToTeam - konwertuje TempTeam na Team (jeśli kompletne)
func (t *TempTeam) ToTeam() (*Team, error) {
	if !t.IsComplete() {
		return nil, fmt.Errorf("drużyna niekompletna, brakujące pola: %v", t.GetMissingFields())
	}
	
	team := &Team{
		Name:      t.Name,
		ShortName: *t.ShortName,
		Name16:    *t.Name16,
		Logo:      *t.Logo,
		Link:      t.Link,      // nullable - przekazujemy jako wskaźnik
		ForeignID: t.ForeignID, // nullable
	}
	
	return team, nil
}