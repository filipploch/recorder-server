package config

import (
	"encoding/json"
	"log"
	"os"
)

// DatabaseConfig - konfiguracja baz danych i rozgrywek
type DatabaseConfig struct {
	CurrentDatabase string                 `json:"current_database"` // ID aktualnych rozgrywek
	DatabasesPath   string                 `json:"databases_path"`
	Competitions    []CompetitionReference `json:"competitions"` // Lista wszystkich rozgrywek
}

// CompetitionReference - referencja do rozgrywek
type CompetitionReference struct {
	ID           string `json:"id"`            // Unikalny ID rozgrywek
	DatabaseFile string `json:"database_file"` // Nazwa pliku bazy danych
	Name         string `json:"name"`          // Pełna nazwa rozgrywek
	PresetID     string `json:"preset_id"`     // ID presetu użytego do utworzenia
	Season       string `json:"season"`        // Sezon np. "2024/2025"
	IsActive     bool   `json:"is_active"`     // Czy rozgrywki są aktywne
}

const DatabaseConfigFile = "database_config.json"

// LoadDatabaseConfig - wczytuje konfigurację bazy danych z pliku
func LoadDatabaseConfig() (*DatabaseConfig, error) {
	// Sprawdź czy plik istnieje
	if _, err := os.Stat(DatabaseConfigFile); os.IsNotExist(err) {
		// NIE twórz domyślnej konfiguracji - zwróć błąd
		log.Println("Database config not found - setup required")
		return nil, err
	}

	// Wczytaj z pliku
	data, err := os.ReadFile(DatabaseConfigFile)
	if err != nil {
		return nil, err
	}

	var config DatabaseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	log.Printf("Database: Loaded config (current: %s, competitions: %d)",
		config.CurrentDatabase, len(config.Competitions))
	return &config, nil
}

// SaveDatabaseConfig - zapisuje konfigurację bazy danych do pliku
func SaveDatabaseConfig(config *DatabaseConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(DatabaseConfigFile, data, 0644); err != nil {
		return err
	}

	log.Printf("Database: Saved config (current: %s)", config.CurrentDatabase)
	return nil
}

// GetCurrentCompetition - zwraca referencję do aktualnych rozgrywek
func (c *DatabaseConfig) GetCurrentCompetition() *CompetitionReference {
	for i := range c.Competitions {
		if c.Competitions[i].ID == c.CurrentDatabase {
			return &c.Competitions[i]
		}
	}
	return nil
}

// AddCompetition - dodaje nowe rozgrywki do konfiguracji
func (c *DatabaseConfig) AddCompetition(comp CompetitionReference) {
	c.Competitions = append(c.Competitions, comp)
}

// SetCurrentCompetition - ustawia aktualne rozgrywki
func (c *DatabaseConfig) SetCurrentCompetition(competitionID string) bool {
	for i := range c.Competitions {
		if c.Competitions[i].ID == competitionID {
			c.CurrentDatabase = competitionID
			return true
		}
	}
	return false
}
