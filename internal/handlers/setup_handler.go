package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"recorder-server/config"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
	"strings"
)

// SetupHandler - handler dla procesu setup
type SetupHandler struct {
	dbManager *database.Manager
}

// NewSetupHandler - tworzy nowy handler setup
func NewSetupHandler(dbManager *database.Manager) *SetupHandler {
	return &SetupHandler{
		dbManager: dbManager,
	}
}

// ShowSetupPage - wyświetla stronę setup
func (h *SetupHandler) ShowSetupPage(w http.ResponseWriter, r *http.Request) {
	// TODO: Renderuj formularz setup
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Setup - Recorder Server</title>
		<style>
			body { font-family: Arial; max-width: 800px; margin: 50px auto; padding: 20px; }
			h1 { color: #333; }
			.btn { padding: 10px 20px; margin: 10px; background: #667eea; color: white; 
			       text-decoration: none; border-radius: 5px; display: inline-block; }
			.btn:hover { background: #5568d3; }
		</style>
	</head>
	<body>
		<h1>🎯 Recorder Server - Konfiguracja</h1>
		<p>Witaj! To pierwsze uruchomienie aplikacji.</p>
		<p>Aby rozpocząć, musisz utworzyć preset rozgrywek, a następnie utworzyć rozgrywki.</p>
		
		<h2>Krok 1: Utwórz preset</h2>
		<a href="/setup/create-preset" class="btn">Utwórz nowy preset</a>
		
		<h2>Krok 2: Utwórz rozgrywki</h2>
		<a href="/setup/create-competition" class="btn">Utwórz rozgrywki z presetu</a>
	</body>
	</html>
	`
	w.Write([]byte(html))
}

// ShowCreatePresetPage - wyświetla formularz tworzenia presetu
func (h *SetupHandler) ShowCreatePresetPage(w http.ResponseWriter, r *http.Request) {
	// TODO: Renderuj formularz tworzenia presetu
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Tworzenie presetu</title>
	</head>
	<body>
		<h1>Tworzenie presetu rozgrywek</h1>
		<form method="POST" action="/setup/create-preset">
			<label>ID presetu: <input type="text" name="id" required></label><br>
			<label>Nazwa: <input type="text" name="name" required></label><br>
			<label>Typ: 
				<select name="competition_type">
					<option value="league">Liga</option>
					<option value="cup">Puchar</option>
					<option value="tournament">Turniej</option>
				</select>
			</label><br>
			<label>Sport: <input type="text" name="sport" value="futsal"></label><br>
			<button type="submit">Utwórz preset</button>
		</form>
	</body>
	</html>
	`
	w.Write([]byte(html))
}

// CreatePreset - tworzy nowy preset
func (h *SetupHandler) CreatePreset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metoda niedozwolona", http.StatusMethodNotAllowed)
		return
	}

	// Wczytaj dane z formularza lub JSON
	var preset config.CompetitionPreset
	
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&preset); err != nil {
			http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
			return
		}
	} else {
		// Formularz HTML
		r.ParseForm()
		preset = config.CompetitionPreset{
			ID:              r.FormValue("id"),
			Name:            r.FormValue("name"),
			CompetitionType: r.FormValue("competition_type"),
			Sport:           r.FormValue("sport"),
		}
	}

	// Wczytaj istniejące presety
	presetsConfig, err := config.LoadPresetsConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Błąd wczytywania presetów: %v", err), http.StatusInternalServerError)
		return
	}

	// Sprawdź czy preset o tym ID już istnieje
	if presetsConfig.GetPresetByID(preset.ID) != nil {
		http.Error(w, "Preset o tym ID już istnieje", http.StatusConflict)
		return
	}

	// Dodaj preset
	presetsConfig.AddPreset(preset)

	// Zapisz
	if err := config.SavePresetsConfig(presetsConfig); err != nil {
		http.Error(w, fmt.Sprintf("Błąd zapisywania presetów: %v", err), http.StatusInternalServerError)
		return
	}

	// Odpowiedź
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Preset utworzony pomyślnie",
		"preset":  preset,
	})
}

// ShowCreateCompetitionPage - wyświetla formularz tworzenia rozgrywek
func (h *SetupHandler) ShowCreateCompetitionPage(w http.ResponseWriter, r *http.Request) {
	// Wczytaj presety
	presetsConfig, err := config.LoadPresetsConfig()
	if err != nil || len(presetsConfig.Presets) == 0 {
		w.Write([]byte("<h1>Brak presetów</h1><p>Najpierw utwórz preset.</p>"))
		return
	}

	// TODO: Renderuj formularz z listą presetów
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Tworzenie rozgrywek</title>
	</head>
	<body>
		<h1>Tworzenie rozgrywek</h1>
		<form method="POST" action="/setup/create-competition">
			<label>Preset: 
				<select name="preset_id" required>
	`
	
	for _, preset := range presetsConfig.Presets {
		html += fmt.Sprintf(`<option value="%s">%s</option>`, preset.ID, preset.Name)
	}
	
	html += `
				</select>
			</label><br>
			<label>ID rozgrywek: <input type="text" name="id" required></label><br>
			<label>Nazwa: <input type="text" name="name" required></label><br>
			<label>Sezon: <input type="text" name="season" value="2024/2025"></label><br>
			<button type="submit">Utwórz rozgrywki</button>
		</form>
	</body>
	</html>
	`
	w.Write([]byte(html))
}

// CreateCompetition - tworzy nowe rozgrywki z presetu
func (h *SetupHandler) CreateCompetition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metoda niedozwolona", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	presetID := r.FormValue("preset_id")
	competitionID := r.FormValue("id")
	competitionName := r.FormValue("name")
	season := r.FormValue("season")

	// Wczytaj preset
	presetsConfig, err := config.LoadPresetsConfig()
	if err != nil {
		http.Error(w, "Błąd wczytywania presetów", http.StatusInternalServerError)
		return
	}

	preset := presetsConfig.GetPresetByID(presetID)
	if preset == nil {
		http.Error(w, "Nie znaleziono presetu", http.StatusNotFound)
		return
	}

	// Nazwa pliku bazy danych
	dbFileName := competitionID + ".db"

	// Utwórz bazę danych
	if err := h.dbManager.CreateDatabase(competitionID); err != nil {
		http.Error(w, fmt.Sprintf("Błąd tworzenia bazy: %v", err), http.StatusInternalServerError)
		return
	}

	// Przełącz na nową bazę
	if err := h.dbManager.SwitchDatabase(competitionID); err != nil {
		http.Error(w, fmt.Sprintf("Błąd przełączania bazy: %v", err), http.StatusInternalServerError)
		return
	}

	// Wykonaj migrację
	if err := h.dbManager.AutoMigrate(models.GetAllModels()...); err != nil {
		http.Error(w, fmt.Sprintf("Błąd migracji: %v", err), http.StatusInternalServerError)
		return
	}

	// Zapisz preset w tabeli Settings
	db := h.dbManager.GetDB()
	presetJSON, _ := json.Marshal(preset)
	settings := models.Settings{
		Name:        "Competition Settings",
		Description: fmt.Sprintf("Settings for %s", competitionName),
		Data:        string(presetJSON),
	}
	db.Create(&settings)

	// Utwórz pusty rekord ActiveSession (singleton)
	activeSession := models.ActiveSession{
		GameID:     nil,
		GamePartID: nil,
	}
	db.Create(&activeSession)

	// Utwórz/zaktualizuj database_config.json
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		// Utwórz nową konfigurację
		dbConfig = &config.DatabaseConfig{
			CurrentDatabase: competitionID,
			DatabasesPath:   "./databases",
			Competitions:    []config.CompetitionReference{},
		}
	}

	// Dodaj rozgrywki
	compRef := config.CompetitionReference{
		ID:           competitionID,
		DatabaseFile: dbFileName,
		Name:         competitionName,
		PresetID:     presetID,
		Season:       season,
		IsActive:     true,
	}
	dbConfig.AddCompetition(compRef)
	dbConfig.SetCurrentCompetition(competitionID)

	// Zapisz konfigurację
	if err := config.SaveDatabaseConfig(dbConfig); err != nil {
		http.Error(w, fmt.Sprintf("Błąd zapisywania konfiguracji: %v", err), http.StatusInternalServerError)
		return
	}

	// Przekieruj na stronę główną
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
