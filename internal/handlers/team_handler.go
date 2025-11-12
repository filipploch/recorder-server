package handlers

import (
	"encoding/json"
	"net/http"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// TeamHandler - handler dla operacji CRUD na zespołach
type TeamHandler struct {
	dbManager *database.Manager
}

// NewTeamHandler - tworzy nowy handler zespołów
func NewTeamHandler(dbManager *database.Manager) *TeamHandler {
	return &TeamHandler{
		dbManager: dbManager,
	}
}

// TeamRequest - struktura żądania dla tworzenia/edycji zespołu
type TeamRequest struct {
	Name      string              `json:"name"`
	ForeignID *string             `json:"foreign_id"`
	ShortName string              `json:"short_name"`
	Name16    string              `json:"name_16"`
	Logo      string              `json:"logo"`
	Link      *string             `json:"link"`
	Kits      map[string][]string `json:"kits"` // Mapa: "1" -> ["#fff"], "2" -> ["#000"], "3" -> ["#66ff73"]
}

// ValidationError - struktura błędu walidacji
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResponse - odpowiedź z błędami walidacji
type ValidationResponse struct {
	Status string            `json:"status"`
	Errors []ValidationError `json:"errors"`
}

// validateTeamRequest - waliduje dane zespołu
func (h *TeamHandler) validateTeamRequest(req TeamRequest, isUpdate bool, teamID uint) []ValidationError {
	errors := []ValidationError{}
	db := h.dbManager.GetDB()

	// Walidacja Name (wymagane, unique, max 255 znaków)
	if strings.TrimSpace(req.Name) == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Nazwa zespołu jest wymagana",
		})
	} else if len(req.Name) > 255 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Nazwa zespołu nie może być dłuższa niż 255 znaków",
		})
	} else {
		// Sprawdź unikalność
		var count int64
		query := db.Model(&models.Team{}).Where("name = ?", req.Name)
		if isUpdate {
			query = query.Where("id != ?", teamID)
		}
		query.Count(&count)
		if count > 0 {
			errors = append(errors, ValidationError{
				Field:   "name",
				Message: "Zespół o tej nazwie już istnieje",
			})
		}
	}

	// Walidacja ShortName (wymagane, unique, dokładnie 3 znaki)
	if strings.TrimSpace(req.ShortName) == "" {
		errors = append(errors, ValidationError{
			Field:   "short_name",
			Message: "Skrót nazwy jest wymagany",
		})
	} else if len(req.ShortName) != 3 {
		errors = append(errors, ValidationError{
			Field:   "short_name",
			Message: "Skrót nazwy musi mieć dokładnie 3 znaki",
		})
	} else {
		// Sprawdź unikalność
		var count int64
		query := db.Model(&models.Team{}).Where("short_name = ?", req.ShortName)
		if isUpdate {
			query = query.Where("id != ?", teamID)
		}
		query.Count(&count)
		if count > 0 {
			errors = append(errors, ValidationError{
				Field:   "short_name",
				Message: "Zespół o tym skrócie już istnieje",
			})
		}
	}

	// Walidacja Name16 (wymagane, unique, max 16 znaków)
	if strings.TrimSpace(req.Name16) == "" {
		errors = append(errors, ValidationError{
			Field:   "name_16",
			Message: "Nazwa 16-znakowa jest wymagana",
		})
	} else if len(req.Name16) > 16 {
		errors = append(errors, ValidationError{
			Field:   "name_16",
			Message: "Nazwa 16-znakowa nie może być dłuższa niż 16 znaków",
		})
	} else {
		// Sprawdź unikalność
		var count int64
		query := db.Model(&models.Team{}).Where("name_16 = ?", req.Name16)
		if isUpdate {
			query = query.Where("id != ?", teamID)
		}
		query.Count(&count)
		if count > 0 {
			errors = append(errors, ValidationError{
				Field:   "name_16",
				Message: "Zespół o tej nazwie 16-znakowej już istnieje",
			})
		}
	}

	// Walidacja Logo (wymagane, URL)
	if strings.TrimSpace(req.Logo) == "" {
		errors = append(errors, ValidationError{
			Field:   "logo",
			Message: "Logo jest wymagane",
		})
	} else if len(req.Logo) > 500 {
		errors = append(errors, ValidationError{
			Field:   "logo",
			Message: "URL logo nie może być dłuższy niż 500 znaków",
		})
	}

	// Walidacja Link (opcjonalne, max 500 znaków)
	if req.Link != nil && len(*req.Link) > 500 {
		errors = append(errors, ValidationError{
			Field:   "link",
			Message: "Link nie może być dłuższy niż 500 znaków",
		})
	}

	// Walidacja ForeignID (opcjonalne, max 255 znaków)
	if req.ForeignID != nil && len(*req.ForeignID) > 255 {
		errors = append(errors, ValidationError{
			Field:   "foreign_id",
			Message: "Foreign ID nie może być dłuższy niż 255 znaków",
		})
	}

	// NOWA WALIDACJA - Kits (jako mapa)
	if len(req.Kits) != 3 {
		errors = append(errors, ValidationError{
			Field:   "kits",
			Message: "Zespół musi mieć dokładnie 3 komplety strojów",
		})
	} else {
		// Sprawdź czy są wszystkie typy (1, 2, 3)
		for kitType := 1; kitType <= 3; kitType++ {
			kitKey := strconv.Itoa(kitType)
			colors, exists := req.Kits[kitKey]

			if !exists {
				errors = append(errors, ValidationError{
					Field:   "kits",
					Message: "Brak kompletu nr " + kitKey,
				})
				continue
			}

			// Sprawdź ilość kolorów (1-5)
			if len(colors) < 1 || len(colors) > 5 {
				errors = append(errors, ValidationError{
					Field:   "kits",
					Message: "Każdy strój musi mieć od 1 do 5 kolorów",
				})
			}

			// Sprawdź format kolorów HEX
			for _, color := range colors {
				if !strings.HasPrefix(color, "#") || len(color) != 7 {
					errors = append(errors, ValidationError{
						Field:   "kits",
						Message: "Nieprawidłowy format koloru (wymagany #RRGGBB)",
					})
					break
				}
			}
		}
	}

	return errors
}

// ListTeams - lista wszystkich zespołów z strojami
func (h *TeamHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	db := h.dbManager.GetDB()

	var teams []models.Team
	// Preload Kits i KitColors
	if err := db.Preload("Kits.KitColors").Order("name ASC").Find(&teams).Error; err != nil {
		http.Error(w, "Błąd pobierania zespołów", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"teams":  teams,
		"count":  len(teams),
	})
}

// GetTeam - pobiera pojedynczy zespół ze strojami
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Nieprawidłowe ID zespołu", http.StatusBadRequest)
		return
	}

	db := h.dbManager.GetDB()
	var team models.Team

	if err := db.Preload("Players").Preload("Coaches").Preload("Kits.KitColors").First(&team, teamID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Zespół nie znaleziony", http.StatusNotFound)
			return
		}
		http.Error(w, "Błąd pobierania zespołu", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"team":   team,
	})
}

// CreateTeam - tworzy nowy zespół ze strojami
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req TeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	// Jeśli nie podano strojów, ustaw domyślne
	if len(req.Kits) == 0 {
		req.Kits = map[string][]string{
			"1": {"#ffffff"},
			"2": {"#000000"},
			"3": {"#66ff73"},
		}
	}

	// Walidacja
	validationErrors := h.validateTeamRequest(req, false, 0)
	if len(validationErrors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ValidationResponse{
			Status: "validation_error",
			Errors: validationErrors,
		})
		return
	}

	db := h.dbManager.GetDB()

	// Rozpocznij transakcję
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Utwórz zespół
	team := models.Team{
		Name:      req.Name,
		ForeignID: req.ForeignID,
		ShortName: req.ShortName,
		Name16:    req.Name16,
		Logo:      req.Logo,
		Link:      req.Link,
	}

	if err := tx.Create(&team).Error; err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Błąd tworzenia zespołu: " + err.Error(),
		})
		return
	}

	// Utwórz stroje z mapy
	for kitType := 1; kitType <= 3; kitType++ {
		kitKey := strconv.Itoa(kitType)
		colors, exists := req.Kits[kitKey]

		if !exists {
			// To nie powinno się zdarzyć po walidacji
			continue
		}

		kit := models.Kit{
			TeamID: team.ID,
			Type:   kitType,
		}

		if err := tx.Create(&kit).Error; err != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "error",
				"error":  "Błąd tworzenia stroju: " + err.Error(),
			})
			return
		}

		// Utwórz kolory stroju
		for order, color := range colors {
			kitColor := models.KitColor{
				KitID:      kit.ID,
				ColorOrder: order + 1, // 1-based index
				Color:      color,
			}

			if err := tx.Create(&kitColor).Error; err != nil {
				tx.Rollback()
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": "error",
					"error":  "Błąd tworzenia koloru stroju: " + err.Error(),
				})
				return
			}
		}
	}

	// Zatwierdź transakcję
	if err := tx.Commit().Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Błąd zatwierdzania transakcji: " + err.Error(),
		})
		return
	}

	// Przeładuj zespół ze strojami
	db.Preload("Kits.KitColors").First(&team, team.ID)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Zespół został utworzony",
		"team":    team,
	})
}

// UpdateTeam - aktualizuje zespół ze strojami
func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Nieprawidłowe ID zespołu", http.StatusBadRequest)
		return
	}

	var req TeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	// Walidacja
	validationErrors := h.validateTeamRequest(req, true, uint(teamID))
	if len(validationErrors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ValidationResponse{
			Status: "validation_error",
			Errors: validationErrors,
		})
		return
	}

	db := h.dbManager.GetDB()

	// Sprawdź czy zespół istnieje i załaduj z Kits
	var team models.Team
	if err := db.Preload("Kits.KitColors").First(&team, teamID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Zespół nie znaleziony", http.StatusNotFound)
			return
		}
		http.Error(w, "Błąd pobierania zespołu", http.StatusInternalServerError)
		return
	}

	// Rozpocznij transakcję
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Aktualizuj dane podstawowe
	team.Name = req.Name
	team.ForeignID = req.ForeignID
	team.ShortName = req.ShortName
	team.Name16 = req.Name16
	team.Logo = req.Logo
	team.Link = req.Link

	if err := tx.Save(&team).Error; err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Błąd aktualizacji zespołu: " + err.Error(),
		})
		return
	}

	// Aktualizuj stroje jeśli podano
	if len(req.Kits) > 0 {
		// Dla każdego typu stroju
		for kitType := 1; kitType <= 3; kitType++ {
			kitKey := strconv.Itoa(kitType)
			colors, exists := req.Kits[kitKey]

			if !exists {
				continue
			}

			// Znajdź odpowiedni Kit dla tego typu
			var kit models.Kit
			found := false
			for _, k := range team.Kits {
				if k.Type == kitType {
					kit = k
					found = true
					break
				}
			}

			if !found {
				// Jeśli Kit nie istnieje (nie powinno się zdarzyć), utwórz go
				kit = models.Kit{
					TeamID: team.ID,
					Type:   kitType,
				}
				if err := tx.Create(&kit).Error; err != nil {
					tx.Rollback()
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "error",
						"error":  "Błąd tworzenia stroju: " + err.Error(),
					})
					return
				}
			}

			// KLUCZOWA ZMIANA: Usuń wszystkie stare kolory dla tego konkretnego Kit
			// Używamy Unscoped() aby fizycznie usunąć rekordy (nie soft delete)
			if err := tx.Unscoped().Where("kit_id = ?", kit.ID).Delete(&models.KitColor{}).Error; err != nil {
				tx.Rollback()
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": "error",
					"error":  "Błąd usuwania starych kolorów: " + err.Error(),
				})
				return
			}

			// Dodaj nowe kolory z poprawnymi ColorOrder (1-based)
			for index, color := range colors {
				kitColor := models.KitColor{
					KitID:      kit.ID,
					ColorOrder: index + 1, // 1-based index
					Color:      color,
				}

				if err := tx.Create(&kitColor).Error; err != nil {
					tx.Rollback()
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "error",
						"error":  "Błąd dodawania koloru: " + err.Error(),
					})
					return
				}
			}
		}
	}

	// Zatwierdź transakcję
	if err := tx.Commit().Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Błąd zatwierdzania transakcji: " + err.Error(),
		})
		return
	}

	// Przeładuj zespół ze strojami
	db.Preload("Kits.KitColors").First(&team, team.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Zespół został zaktualizowany",
		"team":    team,
	})
}

// DeleteTeam - usuwa zespół (pozostaje bez zmian)
func (h *TeamHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Nieprawidłowe ID zespołu", http.StatusBadRequest)
		return
	}

	db := h.dbManager.GetDB()

	// Sprawdź czy zespół istnieje
	var team models.Team
	if err := db.First(&team, teamID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Zespół nie znaleziony", http.StatusNotFound)
			return
		}
		http.Error(w, "Błąd pobierania zespołu", http.StatusInternalServerError)
		return
	}

	// Sprawdź czy zespół jest używany (ma zawodników, trenerów, itp.)
	var playerCount, coachCount int64
	db.Model(&models.Player{}).Where("team_id = ?", teamID).Count(&playerCount)
	db.Model(&models.Coach{}).Where("team_id = ?", teamID).Count(&coachCount)

	if playerCount > 0 || coachCount > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"error":   "Nie można usunąć zespołu, ponieważ ma przypisanych zawodników lub trenerów",
			"players": playerCount,
			"coaches": coachCount,
		})
		return
	}

	// Usuń zespół (soft delete) - cascade usunie Kits i KitColors
	if err := db.Delete(&team).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  "Błąd usuwania zespołu: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Zespół został usunięty",
	})
}
