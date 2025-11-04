package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"recorder-server/config"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
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
		<h1>📊 Recorder Server - Konfiguracja</h1>
		<p>Witaj! To pierwsze uruchomienie aplikacji.</p>
		<p>Aby rozpocząć, musisz utworzyć rozgrywki z dostępnego presetu.</p>
		
		<h2>Utwórz rozgrywki</h2>
		<a href="/setup/create-competition" class="btn">Utwórz rozgrywki z presetu</a>
	</body>
	</html>
	`
	w.Write([]byte(html))
}

// ShowCreateCompetitionPage - wyświetla formularz tworzenia rozgrywek
func (h *SetupHandler) ShowCreateCompetitionPage(w http.ResponseWriter, r *http.Request) {
	// Wczytaj presety
	presetsConfig, err := config.LoadPresetsConfig()
	if err != nil || len(presetsConfig.Presets.Competitions) == 0 {
		w.Write([]byte("<h1>Błąd</h1><p>Brak dostępnych presetów w pliku presets.json lub błąd wczytywania.</p>"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Tworzenie rozgrywek</title>
		<style>
			body { font-family: Arial; max-width: 800px; margin: 50px auto; padding: 20px; }
			h1 { color: #333; }
			form { background: #f5f5f5; padding: 20px; border-radius: 8px; }
			label { display: block; margin: 15px 0 5px; font-weight: bold; }
			input, select, textarea { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
			button { padding: 12px 24px; margin: 20px 0; background: #667eea; color: white; 
			         border: none; border-radius: 5px; cursor: pointer; font-size: 16px; }
			button:hover { background: #5568d3; }
			.help { color: #666; font-size: 14px; margin-top: 5px; }
		</style>
	</head>
	<body>
		<h1>Tworzenie rozgrywek</h1>
		<form method="POST" action="/setup/create-competition">
			<label>Wybierz preset:
				<select name="preset_name" required>
					<option value="">-- Wybierz preset --</option>
	`
	
	for _, preset := range presetsConfig.Presets.Competitions {
		html += fmt.Sprintf(`					<option value="%s">%s</option>`, preset.Name, preset.Name)
		html += "\n"
	}
	
	html += `
				</select>
			</label>
			<p class="help">Preset definiuje reguły i strukturę rozgrywek (liga, puchar, itd.)</p>

			<label>ID rozgrywek:
				<input type="text" name="id" required pattern="[a-zA-Z0-9_-]+" 
				       placeholder="np. nalf_liga_2024_2025">
			</label>
			<p class="help">Unikalny identyfikator (litery, cyfry, myślniki i podkreślenia)</p>

			<label>Nazwa rozgrywek:
				<input type="text" name="name" required placeholder="np. NALF Liga Sezon 2024/2025">
			</label>
			<p class="help">Pełna nazwa rozgrywek widoczna w aplikacji</p>

			<label>Sezon:
				<input type="text" name="season" value="2024/2025" required>
			</label>
			<p class="help">Sezon rozgrywek (np. 2024/2025)</p>

			<label>URL do scrapowania meczów (opcjonalne):
				<input type="url" name="games_url" placeholder="https://example.com/games">
			</label>
			<p class="help">Adres URL skąd pobierane będą informacje o meczach (jeśli dotyczy)</p>

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
	presetName := r.FormValue("preset_name")
	competitionID := r.FormValue("id")
	competitionName := r.FormValue("name")
	season := r.FormValue("season")
	gamesURL := r.FormValue("games_url")

	// Wczytaj presety
	presetsConfig, err := config.LoadPresetsConfig()
	if err != nil {
		http.Error(w, "Błąd wczytywania presetów", http.StatusInternalServerError)
		return
	}

	preset := presetsConfig.GetPresetByName(presetName)
	if preset == nil {
		http.Error(w, "Nie znaleziono presetu", http.StatusNotFound)
		return
	}

	// Utwórz katalog dla rozgrywek
	competitionPath := filepath.Join("./competitions", competitionID)
	if err := os.MkdirAll(competitionPath, 0755); err != nil {
		http.Error(w, fmt.Sprintf("Błąd tworzenia katalogu: %v", err), http.StatusInternalServerError)
		return
	}

	// Nazwa pliku bazy danych
	dbFileName := competitionID + ".db"
	
	// Utwórz bazę przez manager (który utworzy plik w competitions/competitionID/)
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

	db := h.dbManager.GetDB()

	// === ZAPISZ DANE COMMON ===
	
	// TV Staff Roles
	for _, role := range presetsConfig.Presets.Common.TVStaffRoles {
		tvStaffRole := models.TVStaffRole{
			ID:        role.ID,
			Name:      role.Name,
			ShortName: role.ShortName,
		}
		db.Create(&tvStaffRole)
	}

	// Cameras
	for _, cam := range presetsConfig.Presets.Common.Cameras {
		camera := models.Camera{
			ID:       cam.ID,
			Name:     cam.Name,
			Location: cam.Location,
		}
		db.Create(&camera)
	}

	// Coach Roles
	for _, role := range presetsConfig.Presets.Common.CoachRoles {
		coachRole := models.CoachRole{
			ID:        role.ID,
			Name:      role.Name,
			ShortName: role.ShortName,
		}
		db.Create(&coachRole)
	}

	// Referee Roles
	for _, role := range presetsConfig.Presets.Common.RefereeRoles {
		refereeRole := models.RefereeRole{
			ID:        role.ID,
			Name:      role.Name,
			ShortName: role.ShortName,
		}
		db.Create(&refereeRole)
	}

	// === ZAPISZ DANE CONSTANT ===
	
	// Value Types
	for _, vt := range preset.Constant.ValueTypes {
		valueType := models.ValueType{
			ID:   vt.ID,
			Name: vt.Name,
		}
		db.Create(&valueType)
	}

	// Player Roles
	for _, role := range preset.Constant.PlayerRoles {
		playerRole := models.PlayerRole{
			ID:        role.ID,
			Name:      role.Name,
			ShortName: role.ShortName,
		}
		db.Create(&playerRole)
	}

	// === PRZYGOTUJ VARIABLE ===
	
	// Dodaj games_url do variable jeśli podano
	if gamesURL != "" {
		if scraperData, ok := preset.Variable["scraper"].(map[string]interface{}); ok {
			scraperData["games_url"] = gamesURL
		} else {
			preset.Variable["scraper"] = map[string]interface{}{
				"games_url": gamesURL,
			}
		}
	}

	// Serializuj Variable do JSON
	variableJSON, _ := json.Marshal(preset.Variable)

	// Utwórz rekord Competition
	competition := models.Competition{
		Name:     competitionName,
		Season:   season,
		Variable: string(variableJSON),
	}
	if err := db.Create(&competition).Error; err != nil {
		http.Error(w, fmt.Sprintf("Błąd zapisywania rozgrywek: %v", err), http.StatusInternalServerError)
		return
	}

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
			DatabasesPath:   "./competitions",
			Competitions:    []config.CompetitionReference{},
		}
	}

	// Dodaj rozgrywki
	compRef := config.CompetitionReference{
		ID:           competitionID,
		DatabaseFile: dbFileName,
		Name:         competitionName,
		PresetID:     presetName,
		Season:       season,
		IsActive:     true,
	}
	dbConfig.AddCompetition(compRef)
	
	// Ustaw jako aktualną
	if !dbConfig.SetCurrentCompetition(competitionID) {
		http.Error(w, "Błąd ustawiania aktualnej bazy danych", http.StatusInternalServerError)
		return
	}

	// Zapisz konfigurację
	if err := config.SaveDatabaseConfig(dbConfig); err != nil {
		http.Error(w, fmt.Sprintf("Błąd zapisywania konfiguracji: %v", err), http.StatusInternalServerError)
		return
	}

	// Przekieruj na stronę główną
	http.Redirect(w, r, "/", http.StatusSeeOther)
}