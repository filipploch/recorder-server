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

	// Pobierz istniejącą drużynę
	existing, err := manager.Get(tempID)
	if err != nil {
		return err
	}

	// Zachowaj TempID, ScrapedAt
	updated.TempID = existing.TempID
	updated.ScrapedAt = existing.ScrapedAt

	// Jeśli Kits jest nil, zachowaj istniejące lub utwórz puste
	if updated.Kits == nil {
		if existing.Kits != nil {
			updated.Kits = existing.Kits
		} else {
			updated.Kits = map[string][]string{
				"1": {},
				"2": {},
				"3": {},
			}
		}
	}

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

	db := s.dbManager.GetDB()

	// Rozpocznij transakcję
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Zapisz Team do bazy danych
	if err := tx.Create(team).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("błąd zapisu do bazy: %w", err)
	}

	// Utwórz 3 rekordy Kits i ich KitColors
	for kitType := 1; kitType <= 3; kitType++ {
		kit := models.Kit{
			TeamID: team.ID,
			Type:   kitType,
		}

		if err := tx.Create(&kit).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("błąd tworzenia stroju: %w", err)
		}

		// Pobierz kolory dla tego typu stroju z tempTeam.Kits
		kitKey := fmt.Sprintf("%d", kitType)
		colors := []string{}

		if tempTeam.Kits != nil {
			if colorsList, exists := tempTeam.Kits[kitKey]; exists && len(colorsList) > 0 {
				colors = colorsList
			}
		}

		// Jeśli brak kolorów, użyj domyślnych
		if len(colors) == 0 {
			defaultColors := map[int]string{
				1: "#ffffff", // biały dla kompletu 1
				2: "#000000", // czarny dla kompletu 2
				3: "#66ff73", // zielony dla kompletu 3
			}
			colors = []string{defaultColors[kitType]}
		}

		// Utwórz KitColors dla każdego koloru
		for order, color := range colors {
			kitColor := models.KitColor{
				KitID:      kit.ID,
				ColorOrder: order + 1, // 1-based index
				Color:      color,
			}

			if err := tx.Create(&kitColor).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("błąd tworzenia koloru stroju: %w", err)
			}
		}
	}

	// Zatwierdź transakcję
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("błąd zatwierdzania transakcji: %w", err)
	}

	// Usuń z pliku tymczasowego
	if err := manager.Delete(tempID); err != nil {
		// Loguj błąd, ale nie przerywaj - drużyna już jest w bazie
		fmt.Printf("Ostrzeżenie: nie udało się usunąć z pliku tymczasowego: %v\n", err)
	}

	// Przeładuj zespół ze strojami
	db.Preload("Kits.KitColors").First(team, team.ID)

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
