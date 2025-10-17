package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// DatabaseConfig - konfiguracja bazy danych
type DatabaseConfig struct {
	CurrentDatabase string   `json:"current_database"` // Aktualna baza (np. "game_2025_01_15")
	DatabasesPath   string   `json:"databases_path"`   // Ścieżka do katalog z bazami
	AvailableDatabases []string `json:"available_databases"` // Lista dostępnych baz
}

const DatabaseConfigFile = "database_config.json"

// LoadDatabaseConfig - ładuje konfigurację bazy danych z pliku
func LoadDatabaseConfig() (*DatabaseConfig, error) {
	// Sprawdź czy plik istnieje
	if _, err := os.Stat(DatabaseConfigFile); os.IsNotExist(err) {
		// Utwórz domyślną konfigurację
		defaultConfig := &DatabaseConfig{
			CurrentDatabase: "default",
			DatabasesPath:   "./databases",
			AvailableDatabases: []string{"default"},
		}
		
		// Zapisz domyślną konfigurację
		if err := SaveDatabaseConfig(defaultConfig); err != nil {
			return nil, err
		}
		
		log.Println("Database: Utworzono domyślną konfigurację")
		return defaultConfig, nil
	}

	// Wczytaj z pliku
	data, err := ioutil.ReadFile(DatabaseConfigFile)
	if err != nil {
		return nil, err
	}

	var config DatabaseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	log.Printf("Database: Wczytano konfigurację (current: %s)", config.CurrentDatabase)
	return &config, nil
}

// SaveDatabaseConfig - zapisuje konfigurację bazy danych do pliku
func SaveDatabaseConfig(config *DatabaseConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(DatabaseConfigFile, data, 0644); err != nil {
		return err
	}

	log.Printf("Database: Zapisano konfigurację (current: %s)", config.CurrentDatabase)
	return nil
}