package tables

import (
	"errors"
	"recorder-server/internal/models"
)

// TeamStanding - pozycja drużyny w tabeli
type TeamStanding struct {
	Position       int           `json:"position"`        // Pozycja w tabeli
	Team           *models.Team  `json:"team"`            // Drużyna
	TeamID         uint          `json:"team_id"`         // ID drużyny
	Played         int           `json:"played"`          // Rozegrane mecze
	Won            int           `json:"won"`             // Wygrane
	Drawn          int           `json:"drawn"`           // Remisy
	Lost           int           `json:"lost"`            // Przegrane
	GoalsFor       int           `json:"goals_for"`       // Bramki zdobyte
	GoalsAgainst   int           `json:"goals_against"`   // Bramki stracone
	GoalDifference int           `json:"goal_difference"` // Różnica bramek
	Points         int           `json:"points"`          // Punkty
	Form           string        `json:"form"`            // Forma (np. "WWLDW")
	CustomData     map[string]interface{} `json:"custom_data,omitempty"` // Dodatkowe dane specyficzne dla algorytmu
}

// Table - kompletna tabela
type Table struct {
	GroupID    uint            `json:"group_id"`
	GroupName  string          `json:"group_name"`
	Standings  []TeamStanding  `json:"standings"`
	UpdatedAt  string          `json:"updated_at"`
	Algorithm  string          `json:"algorithm"` // Użyty algorytm
}

// TableOrderAlgorithm - interfejs dla algorytmów sortowania tabeli
type TableOrderAlgorithm interface {
	GetName() string // Nazwa algorytmu
	
	// CalculateTable - oblicza tabelę dla danej grupy
	CalculateTable(groupID uint, teams []models.Team, games []models.Game) (*Table, error)
	
	// CompareTeams - porównuje dwie drużyny i zwraca:
	// -1 jeśli team1 powinna być wyżej
	//  0 jeśli są równe
	// +1 jeśli team2 powinna być wyżej
	CompareTeams(standing1, standing2 *TeamStanding, headToHeadGames []models.Game) int
}

// Errors
var (
	ErrAlgorithmNotFound = errors.New("table order algorithm not found")
	ErrInvalidGroupData  = errors.New("invalid group data")
	ErrCalculationFailed = errors.New("table calculation failed")
)
