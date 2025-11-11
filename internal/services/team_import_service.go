package services

import (
	"fmt"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
)

// TeamImportService - serwis do importu drużyn z pliku tymczasowego
type TeamImportService struct {
	dbManager *database.Manager
}

// NewTeamImportService - tworzy nowy serwis importu
func NewTeamImportService(dbManager *database.Manager) *TeamImportService {
	return &TeamImportService{
		dbManager: dbManager,
	}
}

// GetTempTeams - pobiera wszystkie tymczasowe drużyny
func (s *TeamImportService) GetTempTeams(competitionID string) (*models.TempTeamsCollection, error) {
	manager := models.NewTempTeamManager(competitionID)
	return manager.Load()
}

// GetTempTeam - pobiera pojedynczą tymczasową drużynę
func (s *TeamImportService) GetTempTeam(competitionID, tempID string) (*models.TempTeam, error) {
	manager := models.NewTempTeamManager(competitionID)
	return manager.Get(tempID)
}

// UpdateTempTeam - aktualizuje tymczasową drużynę
func (s *TeamImportService) UpdateTempTeam(competitionID string, tempID string, updated models.TempTeam) error {
	manager := models.NewTempTeamManager(competitionID)
	return manager.Update(tempID, updated)
}

// ImportTeam - importuje drużynę z pliku tymczasowego do bazy danych
func (s *TeamImportService) ImportTeam(competitionID, tempID string) (*models.Team, error) {
	manager := models.NewTempTeamManager(competitionID)
	
	// Pobierz tymczasową drużynę
	tempTeam, err := manager.Get(tempID)
	if err != nil {
		return nil, fmt.Errorf("błąd pobierania tymczasowej drużyny: %w", err)
	}
	
	// Konwertuj na Team
	team, err := tempTeam.ToTeam()
	if err != nil {
		return nil, fmt.Errorf("błąd konwersji: %w", err)
	}
	
	// Zapisz do bazy danych
	db := s.dbManager.GetDB()
	if err := db.Create(team).Error; err != nil {
		return nil, fmt.Errorf("błąd zapisu do bazy: %w", err)
	}
	
	// Usuń z pliku tymczasowego
	if err := manager.Delete(tempID); err != nil {
		// Loguj błąd, ale nie przerywaj - drużyna już jest w bazie
		fmt.Printf("Ostrzeżenie: nie udało się usunąć z pliku tymczasowego: %v\n", err)
	}
	
	return team, nil
}

// ImportAllComplete - importuje wszystkie kompletne drużyny
func (s *TeamImportService) ImportAllComplete(competitionID string) (imported int, errors []error) {
	manager := models.NewTempTeamManager(competitionID)
	collection, err := manager.Load()
	if err != nil {
		return 0, []error{err}
	}
	
	db := s.dbManager.GetDB()
	imported = 0
	errors = []error{}
	
	for _, tempTeam := range collection.Teams {
		if !tempTeam.IsComplete() {
			continue
		}
		
		team, err := tempTeam.ToTeam()
		if err != nil {
			errors = append(errors, fmt.Errorf("drużyna %s: %w", tempTeam.TempID, err))
			continue
		}
		
		if err := db.Create(team).Error; err != nil {
			errors = append(errors, fmt.Errorf("drużyna %s: %w", tempTeam.TempID, err))
			continue
		}
		
		// Usuń z pliku tymczasowego
		if err := manager.Delete(tempTeam.TempID); err != nil {
			// Tylko loguj - drużyna jest już w bazie
			fmt.Printf("Ostrzeżenie: nie udało się usunąć %s z pliku tymczasowego: %v\n", 
				tempTeam.TempID, err)
		}
		
		imported++
	}
	
	return imported, errors
}

// DeleteTempTeam - usuwa tymczasową drużynę (rezygnacja z importu)
func (s *TeamImportService) DeleteTempTeam(competitionID, tempID string) error {
	manager := models.NewTempTeamManager(competitionID)
	return manager.Delete(tempID)
}

// GetStatistics - zwraca statystyki tymczasowych drużyn
func (s *TeamImportService) GetStatistics(competitionID string) (map[string]interface{}, error) {
	manager := models.NewTempTeamManager(competitionID)
	collection, err := manager.Load()
	if err != nil {
		return nil, err
	}
	
	complete := 0
	incomplete := 0
	
	for _, team := range collection.Teams {
		if team.IsComplete() {
			complete++
		} else {
			incomplete++
		}
	}
	
	return map[string]interface{}{
		"total":      len(collection.Teams),
		"complete":   complete,
		"incomplete": incomplete,
	}, nil
}

// ValidateBeforeImport - waliduje drużynę przed importem
func (s *TeamImportService) ValidateBeforeImport(competitionID, tempID string) (bool, []string, error) {
	manager := models.NewTempTeamManager(competitionID)
	tempTeam, err := manager.Get(tempID)
	if err != nil {
		return false, nil, err
	}
	
	missing := tempTeam.GetMissingFields()
	isValid := len(missing) == 0
	
	// Dodatkowa walidacja - sprawdź czy Name, ShortName i Name16 nie są już w bazie
	if isValid {
		db := s.dbManager.GetDB()
		
		// Sprawdź Name (nowe)
		var count int64
		db.Model(&models.Team{}).Where("name = ?", tempTeam.Name).Count(&count)
		if count > 0 {
			missing = append(missing, "name already exists")
			isValid = false
		}
		
		// Sprawdź ShortName
		db.Model(&models.Team{}).Where("short_name = ?", *tempTeam.ShortName).Count(&count)
		if count > 0 {
			missing = append(missing, "short_name already exists")
			isValid = false
		}
		
		// Sprawdź Name16
		db.Model(&models.Team{}).Where("name_16 = ?", *tempTeam.Name16).Count(&count)
		if count > 0 {
			missing = append(missing, "name_16 already exists")
			isValid = false
		}
	}
	
	return isValid, missing, nil
}
