package services

import (
	"encoding/json"
	"fmt"
	"log"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
	"recorder-server/internal/tables"
)

// TableService - serwis obsługujący operacje na tabelach
type TableService struct {
	dbManager *database.Manager
	registry  *tables.AlgorithmRegistry
}

// NewTableService - tworzy nowy serwis tabel
func NewTableService(dbManager *database.Manager) *TableService {
	return &TableService{
		dbManager: dbManager,
		registry:  tables.GetAlgorithmRegistry(),
	}
}

// getAlgorithmNameFromVariable - pobiera nazwę algorytmu z pola Variable
func (s *TableService) getAlgorithmNameFromVariable(variable string) (string, error) {
	var variableData map[string]interface{}
	if err := json.Unmarshal([]byte(variable), &variableData); err != nil {
		return "", fmt.Errorf("błąd parsowania Variable: %w", err)
	}
	
	// Sprawdź czy istnieje pole table_order_algorithm
	if algorithm, ok := variableData["table_order_algorithm"].(string); ok && algorithm != "" {
		return algorithm, nil
	}
	
	return "", fmt.Errorf("brak informacji o algorytmie sortowania w Variable")
}

// CalculateTableForGroup - oblicza tabelę dla danej grupy
func (s *TableService) CalculateTableForGroup(groupID uint) (*tables.Table, error) {
	db := s.dbManager.GetDB()
	
	// Pobierz grupę z relacjami
	var group models.Group
	if err := db.Preload("Stage.Competition").Preload("GroupTeams.Team").First(&group, groupID).Error; err != nil {
		return nil, fmt.Errorf("nie znaleziono group: %w", err)
	}
	
	competition := group.Stage.Competition
	
	// Pobierz nazwę algorytmu z Variable
	algorithmName, err := s.getAlgorithmNameFromVariable(competition.Variable)
	if err != nil {
		return nil, fmt.Errorf("brak przypisanego algorytmu sortowania dla group ID=%d: %w", groupID, err)
	}
	
	log.Printf("TableService: Obliczanie tabeli dla grupy '%s' używając algorytmu '%s'",
		group.Name, algorithmName)
	
	// Pobierz algorytm z registry
	algorithm, err := s.registry.GetAlgorithm(algorithmName)
	if err != nil {
		return nil, fmt.Errorf("błąd pobierania algorytmu: %w", err)
	}
	
	// Pobierz drużyny w grupie
	teams := make([]models.Team, 0, len(group.GroupTeams))
	for _, gt := range group.GroupTeams {
		teams = append(teams, gt.Team)
	}
	
	// Pobierz mecze dla tej grupy
	var games []models.Game
	if err := db.Where("group_id = ?", groupID).Find(&games).Error; err != nil {
		return nil, fmt.Errorf("błąd pobierania meczów: %w", err)
	}
	
	// Oblicz tabelę używając algorytmu
	log.Printf("TableService: Uruchamianie algorytmu '%s'...", algorithm.GetName())
	table, err := algorithm.CalculateTable(groupID, teams, games)
	if err != nil {
		return nil, fmt.Errorf("błąd obliczania tabeli: %w", err)
	}
	
	// Dodaj nazwę grupy
	table.GroupName = group.Name
	
	log.Printf("TableService: Obliczono tabelę z %d pozycjami", len(table.Standings))
	return table, nil
}

// CalculateTableForStage - oblicza tabele dla wszystkich grup w stage
func (s *TableService) CalculateTableForStage(stageID uint) ([]*tables.Table, error) {
	db := s.dbManager.GetDB()
	
	// Pobierz stage z grupami
	var stage models.Stage
	if err := db.Preload("Groups").First(&stage, stageID).Error; err != nil {
		return nil, fmt.Errorf("nie znaleziono stage: %w", err)
	}
	
	log.Printf("TableService: Obliczanie tabel dla stage '%s' (%d grup)",
		stage.Name, len(stage.Groups))
	
	// Oblicz tabelę dla każdej grupy
	result := make([]*tables.Table, 0, len(stage.Groups))
	for _, group := range stage.Groups {
		table, err := s.CalculateTableForGroup(group.ID)
		if err != nil {
			log.Printf("TableService: Błąd dla grupy %d: %v", group.ID, err)
			continue // Kontynuuj z następną grupą
		}
		result = append(result, table)
	}
	
	log.Printf("TableService: Obliczono %d tabel", len(result))
	return result, nil
}

// CalculateTableForCompetition - oblicza tabele dla wszystkich grup w competition
func (s *TableService) CalculateTableForCompetition(competitionID uint) ([]*tables.Table, error) {
	db := s.dbManager.GetDB()
	
	// Pobierz competition z stages
	var competition models.Competition
	if err := db.Preload("Stages").First(&competition, competitionID).Error; err != nil {
		return nil, fmt.Errorf("nie znaleziono competition: %w", err)
	}
	
	log.Printf("TableService: Obliczanie tabel dla competition '%s' (%d stages)",
		competition.Name, len(competition.Stages))
	
	// Oblicz tabele dla każdego stage
	allTables := make([]*tables.Table, 0)
	for _, stage := range competition.Stages {
		stageTables, err := s.CalculateTableForStage(stage.ID)
		if err != nil {
			log.Printf("TableService: Błąd dla stage %d: %v", stage.ID, err)
			continue
		}
		allTables = append(allTables, stageTables...)
	}
	
	log.Printf("TableService: Obliczono łącznie %d tabel", len(allTables))
	return allTables, nil
}

// GetAvailableAlgorithms - zwraca listę dostępnych algorytmów
func (s *TableService) GetAvailableAlgorithms() []string {
	return s.registry.ListAlgorithms()
}

// CompareTeamsInGroup - porównuje dwie drużyny w kontekście grupy (dla celów debugowania)
func (s *TableService) CompareTeamsInGroup(groupID uint, team1ID, team2ID uint) (int, error) {
	db := s.dbManager.GetDB()
	
	// Pobierz grupę z competition
	var group models.Group
	if err := db.Preload("Stage.Competition").First(&group, groupID).Error; err != nil {
		return 0, fmt.Errorf("nie znaleziono group: %w", err)
	}
	
	competition := group.Stage.Competition
	
	// Pobierz nazwę algorytmu z Variable
	algorithmName, err := s.getAlgorithmNameFromVariable(competition.Variable)
	if err != nil {
		return 0, fmt.Errorf("brak przypisanego algorytmu: %w", err)
	}
	
	// Pobierz algorytm
	algorithm, err := s.registry.GetAlgorithm(algorithmName)
	if err != nil {
		return 0, err
	}
	
	// TODO: Pobierz statystyki drużyn i mecze bezpośrednie
	// Dla uproszczenia zwracamy 0
	
	log.Printf("TableService: Porównanie drużyn %d i %d używając algorytmu '%s'",
		team1ID, team2ID, algorithm.GetName())
	
	return 0, nil
}