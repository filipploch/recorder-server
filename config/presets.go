package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// PresetsConfig - konfiguracja presetów rozgrywek
type PresetsConfig struct {
	Presets []CompetitionPreset `json:"presets"`
}

// CompetitionPreset - szablon/preset rozgrywek
type CompetitionPreset struct {
	ID                string            `json:"id"`                  // Unikalny ID presetu
	Name              string            `json:"name"`                // Nazwa presetu
	CompetitionType   string            `json:"competition_type"`    // "league", "cup", "tournament"
	Sport             string            `json:"sport"`               // "futsal", "basketball", etc.
	TeamsCount        *int              `json:"teams_count"`         // nullable - ilość drużyn
	IsTwoLeg          bool              `json:"is_two_leg"`          // Czy dwumecze
	MatchesPerPairing int               `json:"matches_per_pairing"` // Ile meczy każda para drużyn
	ValueTypes        []string          `json:"value_types"`         // Lista typów wartości (goals, fouls)
	GameParts         []GamePartPreset  `json:"game_parts"`          // Części meczu
	PlayerRoles       []string          `json:"player_roles"`        // Role zawodników
	CoachRoles        []string          `json:"coach_roles"`         // Role trenerów
	RefereeRoles      []string          `json:"referee_roles"`       // Role sędziów
	TVStaffRoles      []string          `json:"tv_staff_roles"`      // Role ekipy TV
	KitTypes          []string          `json:"kit_types"`           // Typy strojów
	EventTypes        []EventTypePreset `json:"event_types"`         // Typy wydarzeń
}

// GamePartPreset - szablon części meczu
type GamePartPreset struct {
	Name      string `json:"name"`       // Nazwa części
	Length    *int   `json:"length"`     // nullable - długość w sekundach
	TimeGroup *int   `json:"time_group"` // nullable - grupa sumowania czasu
}

// EventTypePreset - szablon typu wydarzenia
type EventTypePreset struct {
	Name         string `json:"name"`
	ShortName    string `json:"short_name"`
	Icon         string `json:"icon"`
	IsInProtocol bool   `json:"is_in_protocol"`
}

const PresetsConfigFile = "presets.json"

// LoadPresetsConfig - wczytuje konfigurację presetów z pliku
func LoadPresetsConfig() (*PresetsConfig, error) {
	// Sprawdź czy plik istnieje
	if _, err := os.Stat(PresetsConfigFile); os.IsNotExist(err) {
		// Utwórz pusty plik z presetami
		emptyConfig := &PresetsConfig{
			Presets: []CompetitionPreset{},
		}
		
		if err := SavePresetsConfig(emptyConfig); err != nil {
			return nil, err
		}
		
		log.Println("Presets: Created empty presets file")
		return emptyConfig, nil
	}

	// Wczytaj z pliku
	data, err := ioutil.ReadFile(PresetsConfigFile)
	if err != nil {
		return nil, err
	}

	var config PresetsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	log.Printf("Presets: Loaded %d presets", len(config.Presets))
	return &config, nil
}

// SavePresetsConfig - zapisuje konfigurację presetów do pliku
func SavePresetsConfig(config *PresetsConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(PresetsConfigFile, data, 0644); err != nil {
		return err
	}

	log.Printf("Presets: Saved %d presets", len(config.Presets))
	return nil
}

// GetPresetByID - zwraca preset o podanym ID
func (c *PresetsConfig) GetPresetByID(presetID string) *CompetitionPreset {
	for i := range c.Presets {
		if c.Presets[i].ID == presetID {
			return &c.Presets[i]
		}
	}
	return nil
}

// AddPreset - dodaje nowy preset
func (c *PresetsConfig) AddPreset(preset CompetitionPreset) {
	c.Presets = append(c.Presets, preset)
}

// UpdatePreset - aktualizuje istniejący preset
func (c *PresetsConfig) UpdatePreset(preset CompetitionPreset) bool {
	for i := range c.Presets {
		if c.Presets[i].ID == preset.ID {
			c.Presets[i] = preset
			return true
		}
	}
	return false
}

// DeletePreset - usuwa preset
func (c *PresetsConfig) DeletePreset(presetID string) bool {
	for i := range c.Presets {
		if c.Presets[i].ID == presetID {
			c.Presets = append(c.Presets[:i], c.Presets[i+1:]...)
			return true
		}
	}
	return false
}
