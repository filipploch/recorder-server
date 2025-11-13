package config

import (
	"encoding/json"
	"log"
	"os"
)

// PresetsConfig - konfiguracja presetów rozgrywek
type PresetsConfig struct {
	Presets PresetsData `json:"presets"`
}

// PresetsData - struktura głównych danych presetów
type PresetsData struct {
	Competitions []CompetitionPresetFull `json:"competitions"`
	Common       CommonData              `json:"common"`
}

// CompetitionPresetFull - pełny preset rozgrywek
type CompetitionPresetFull struct {
	Name     string                 `json:"name"`
	Constant ConstantData           `json:"constant"`
	Variable map[string]interface{} `json:"variable"`
}

// ConstantData - dane stałe dla rozgrywek
type ConstantData struct {
	ValueTypes     []ValueTypeData     `json:"value_types"`
	PlayerRoles    []RoleData          `json:"player_roles"`
	GamePartValues []GamePartValueData `json:"game_part_values"`
	Stages         []StageData         `json:"stages"`
	Groups         []GroupData         `json:"groups"`
}

// CommonData - dane wspólne dla wszystkich rozgrywek
type CommonData struct {
	TVStaffRoles []RoleData   `json:"tv_staff_roles"`
	Cameras      []CameraData `json:"cameras"`
	CoachRoles   []RoleData   `json:"coach_roles"`
	RefereeRoles []RoleData   `json:"referee_roles"`
}

// ValueTypeData - typ wartości (bramki, faule)
type ValueTypeData struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// RoleData - uniwersalna struktura dla ról
type RoleData struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name,omitempty"`
	Location  string `json:"location,omitempty"` // dla kamer
}

// CameraData - dane kamery
type CameraData struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

// GamePartValueData - wartość części meczu
type GamePartValueData struct {
	ValueTypeID    uint `json:"value_type_id"`
	Value          *int `json:"value"`            // nullable
	GameValueGroup *int `json:"game_value_group"` // nullable
	MinValue       *int `json:"min_value"`        // nullable
	MaxValue       *int `json:"max_value"`        // nullable
}

// StagesData - dane etapów rozgrywek
type StageData struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	StageOrder     uint    `json:"stage_order"`
	PromotionRules *string `json:"promotion_rules"` // nullable
}

// StagesData - dane etapów rozgrywek
type GroupData struct {
	ID                     uint    `json:"id"`
	StageID                uint    `json:"stage_id"`
	Name                   string  `json:"name"`
	SpecificPromotionRules *string `json:"specific_promotion_rules"` // nullable
	NumberInStage          uint    `json:"number_in_stage"`
}

const PresetsConfigFile = "presets.json"

// LoadPresetsConfig - wczytuje konfigurację presetów z pliku
func LoadPresetsConfig() (*PresetsConfig, error) {
	// Sprawdź czy plik istnieje
	if _, err := os.Stat(PresetsConfigFile); os.IsNotExist(err) {
		return nil, err
	}

	// Wczytaj z pliku
	data, err := os.ReadFile(PresetsConfigFile)
	if err != nil {
		return nil, err
	}

	var config PresetsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	log.Printf("Presets: Loaded %d competition presets", len(config.Presets.Competitions))
	return &config, nil
}

// GetPresetByName - zwraca preset o podanej nazwie
func (c *PresetsConfig) GetPresetByName(name string) *CompetitionPresetFull {
	for i := range c.Presets.Competitions {
		if c.Presets.Competitions[i].Name == name {
			return &c.Presets.Competitions[i]
		}
	}
	return nil
}

// GetPresetNames - zwraca listę nazw presetów
func (c *PresetsConfig) GetPresetNames() []string {
	names := make([]string, len(c.Presets.Competitions))
	for i, preset := range c.Presets.Competitions {
		names[i] = preset.Name
	}
	return names
}
