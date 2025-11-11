package main

import (
	"encoding/json"
	// "fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"recorder-server/config"
	"recorder-server/internal/database"
	"recorder-server/internal/handlers"
	"recorder-server/internal/models"
	"recorder-server/internal/scrapers"
	"recorder-server/internal/services"
	"recorder-server/internal/state"
	"recorder-server/internal/tables"
	"strings"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("=== Recorder Server ===")
	log.Println("Inicjalizacja aplikacji...")

	// Załaduj konfigurację
	cfg := config.LoadConfig()
	log.Printf("Konfiguracja załadowana: Port=%s, OBS URL=%s", cfg.Server.Port, cfg.OBS.URL)

	// Inicjalizacja Database Manager
	dbManager := database.GetManager()

	// Sprawdź czy istnieje konfiguracja bazy danych
	dbConfig, err := config.LoadDatabaseConfig()
	needsSetup := err != nil

	if !needsSetup {
		// Konfiguracja istnieje - inicjalizuj normalnie
		if err := dbManager.Initialize(); err != nil {
			log.Println("Ostrzeżenie: Błąd inicjalizacji Database Manager:", err)
			needsSetup = true
		} else {
			log.Println("Database Manager zainicjalizowany")

			// Wykonaj migrację dla aktualnej bazy
			log.Println("Wykonywanie migracji bazy danych...")
			if err := dbManager.AutoMigrate(models.GetAllModels()...); err != nil {
				log.Printf("OSTRZEŻENIE: Błąd migracji bazy danych: %v", err)
			} else {
				log.Printf("Migracja zakończona pomyślnie dla bazy: %s", dbManager.GetCurrentDatabaseName())
			}

			// Wczytaj aktywną sesję
			db := dbManager.GetDB()
			var activeSession models.ActiveSession
			if err := db.Preload("Game").Preload("GamePart").First(&activeSession).Error; err == nil {
				if activeSession.GameID != nil {
					log.Printf("Aktywna sesja: Game ID=%d", *activeSession.GameID)
					if activeSession.GamePartID != nil {
						log.Printf("Aktywna część meczu: GamePart ID=%d", *activeSession.GamePartID)
					}
				} else {
					log.Println("Brak aktywnej sesji nagrywania")
				}
			}
		}
	}

	defer dbManager.Close()

	// ===== Inicjalizacja scraperów =====
	log.Println("Inicjalizacja scraperów...")
	scrapers.RegisterExampleScrapers()

	// ===== Inicjalizacja algorytmów tabel =====
	log.Println("Inicjalizacja algorytmów tabel...")
	tables.RegisterDefaultAlgorithms()

	// Inicjalizacja stanu aplikacji
	appState := state.NewAppState(cfg.Recording.AllCameras)
	log.Println("Stan aplikacji zainicjalizowany")

	// Inicjalizacja serwisu Socket.IO
	socketService := services.NewSocketIOService(appState)
	log.Println("Socket.IO serwis zainicjalizowany")

	// Inicjalizacja serwisu stopera
	timerService := services.NewTimerService(socketService)
	log.Println("Timer serwis zainicjalizowany")

	// Inicjalizacja klienta OBS WebSocket
	obsClient := services.NewOBSClient(cfg.OBS.URL, cfg.OBS.Password)
	go obsClient.Connect()
	setupOBSEventHandlers(obsClient)
	log.Println("OBS WebSocket klient zainicjalizowany")

	// ===== Inicjalizacja serwisów =====
	scraperService := services.NewScraperService(dbManager)
	// tableService := services.NewTableService(dbManager)

	// Inicjalizacja handlerów
	setupHandler := handlers.NewSetupHandler(dbManager)
	sessionHandler := handlers.NewSessionHandler(dbManager)
	pageHandler := handlers.NewPageHandler()
	cameraHandler := handlers.NewCameraHandler(appState, socketService)
	obsHandler := handlers.NewOBSHandler(obsClient)
	timerHandler := handlers.NewTimerHandler(timerService)
	databaseHandler := handlers.NewDatabaseHandler(dbManager)
	scraperHandler := handlers.NewScraperHandler(dbManager) // Przekaż dbManager
	// tableHandler := handlers.NewTableHandler(tableService)
	teamHandler := handlers.NewTeamHandler(dbManager)
	logoHandler := handlers.NewLogoHandler()
	teamImportHandler := handlers.NewTeamImportHandler(dbManager)
	log.Println("Handlery HTTP zainicjalizowane")

	// Router
	router := mux.NewRouter()

	// Socket.IO - MUSI BYĆ PRZED MIDDLEWARE
	router.Handle("/socket.io/", socketService.GetServer())

	// Setup routes - BEZ middleware (zawsze dostępne)
	setupRouter := router.PathPrefix("/setup").Subrouter()
	setupRouter.HandleFunc("", setupHandler.ShowSetupPage).Methods("GET")
	setupRouter.HandleFunc("/", setupHandler.ShowSetupPage).Methods("GET")
	setupRouter.HandleFunc("/create-competition", setupHandler.ShowCreateCompetitionPage).Methods("GET")
	setupRouter.HandleFunc("/create-competition", setupHandler.CreateCompetition).Methods("POST")

	// Pliki statyczne - BEZ middleware
	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))),
	)

	// Middleware sprawdzający konfigurację
	router.Use(checkSetupMiddleware(dbConfig))

	// Strony WWW
	router.HandleFunc("/", pageHandler.Index).Methods("GET")

	// Strona zarządzania zespołami
	router.HandleFunc("/teams", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("web/templates/teams.html"))
		tmpl.Execute(w, nil)
	}).Methods("GET")

	// API - Kamery
	router.HandleFunc("/api/start-recording", cameraHandler.StartRecording).Methods("POST")
	router.HandleFunc("/api/stop-recording", cameraHandler.StopRecording).Methods("POST")
	router.HandleFunc("/api/get-record-data", cameraHandler.GetRecordData).Methods("POST")
	router.HandleFunc("/api/status", cameraHandler.GetStatus).Methods("GET")

	// API - OBS
	router.HandleFunc("/api/obs/start-recording", obsHandler.StartRecording).Methods("POST")
	router.HandleFunc("/api/obs/stop-recording", obsHandler.StopRecording).Methods("POST")
	router.HandleFunc("/api/obs/status", obsHandler.GetStatus).Methods("GET")
	router.HandleFunc("/api/obs/scenes", obsHandler.GetScenes).Methods("GET")
	router.HandleFunc("/api/obs/set-scene", obsHandler.SetScene).Methods("POST")

	// API - Timer
	router.HandleFunc("/api/timer/start", timerHandler.Start).Methods("POST")
	router.HandleFunc("/api/timer/pause", timerHandler.Pause).Methods("POST")
	router.HandleFunc("/api/timer/reset", timerHandler.Reset).Methods("POST")
	router.HandleFunc("/api/timer/state", timerHandler.GetState).Methods("GET")

	// API - Database
	router.HandleFunc("/api/database/current", databaseHandler.GetCurrent).Methods("GET")
	router.HandleFunc("/api/database/available", databaseHandler.GetAvailable).Methods("GET")
	router.HandleFunc("/api/database/switch", databaseHandler.SwitchDatabase).Methods("POST")
	router.HandleFunc("/api/database/create", databaseHandler.CreateDatabase).Methods("POST")
	router.HandleFunc("/api/database/delete", databaseHandler.DeleteDatabase).Methods("DELETE")

	// API - Session
	router.HandleFunc("/api/session/current", sessionHandler.GetActiveSession).Methods("GET")
	router.HandleFunc("/api/session/set-game", sessionHandler.SetActiveGame).Methods("POST")
	router.HandleFunc("/api/session/set-gamepart", sessionHandler.SetActiveGamePart).Methods("POST")
	router.HandleFunc("/api/session/clear", sessionHandler.ClearActiveSession).Methods("POST")

	// API - Scrapers
	router.HandleFunc("/api/scrape/teams", scraperHandler.ScrapeTeams).Methods("POST")
	router.HandleFunc("/api/scrape/players", scraperHandler.ScrapePlayers).Methods("POST")
	router.HandleFunc("/api/scrape/games", scraperHandler.ScrapeGames).Methods("POST")
	router.HandleFunc("/api/scrape/available", scraperHandler.GetAvailableScrapers).Methods("GET")
	router.HandleFunc("/api/scrape/competition/info", scraperHandler.GetCompetitionScraperInfo).Methods("GET")

	// API - Tables
	// router.HandleFunc("/api/tables/group", tableHandler.CalculateTableForGroup).Methods("GET")
	// router.HandleFunc("/api/tables/stage", tableHandler.CalculateTableForStage).Methods("GET")
	// router.HandleFunc("/api/tables/competition", tableHandler.CalculateTableForCompetition).Methods("GET")
	// router.HandleFunc("/api/tables/algorithms", tableHandler.GetAvailableAlgorithms).Methods("GET")
	// router.HandleFunc("/api/tables/competition/algorithm", tableHandler.GetCompetitionAlgorithmInfo).Methods("GET")
	// router.HandleFunc("/api/tables/compare", tableHandler.CompareTeams).Methods("GET")

	// API - Teams CRUD
	router.HandleFunc("/api/teams", teamHandler.ListTeams).Methods("GET")
	router.HandleFunc("/api/teams", teamHandler.CreateTeam).Methods("POST")
	router.HandleFunc("/api/teams/{id}", teamHandler.GetTeam).Methods("GET")
	router.HandleFunc("/api/teams/{id}", teamHandler.UpdateTeam).Methods("PUT")
	router.HandleFunc("/api/teams/{id}", teamHandler.DeleteTeam).Methods("DELETE")

	// API - Team Import (tymczasowe drużyny)
	// teamImportHandler := handlers.NewTeamImportHandler(dbManager)
	// teamImportHandler = handlers.NewTeamImportHandler(dbManager)
	// router.HandleFunc("/api/teams/temp", teamImportHandler.GetTempTeams).Methods("GET")
	// router.HandleFunc("/api/teams/temp/{temp_id}", teamImportHandler.GetTempTeam).Methods("GET")
	// router.HandleFunc("/api/teams/temp/{temp_id}", teamImportHandler.UpdateTempTeam).Methods("PUT")
	// router.HandleFunc("/api/teams/temp/{temp_id}", teamImportHandler.DeleteTempTeam).Methods("DELETE")
	// router.HandleFunc("/api/teams/import/{temp_id}", teamImportHandler.ImportTeam).Methods("POST")
	// router.HandleFunc("/api/teams/import-all", teamImportHandler.ImportAllComplete).Methods("POST")

	// API - Logos
	router.HandleFunc("/api/logos", logoHandler.ListLogos).Methods("GET")

	// API - Team Import (tymczasowe drużyny)
	router.HandleFunc("/api/teams/temp", teamImportHandler.GetTempTeams).Methods("GET")
	router.HandleFunc("/api/teams/temp/{temp_id}", teamImportHandler.GetTempTeam).Methods("GET")
	router.HandleFunc("/api/teams/temp/{temp_id}", teamImportHandler.UpdateTempTeam).Methods("PUT")
	router.HandleFunc("/api/teams/temp/{temp_id}", teamImportHandler.DeleteTempTeam).Methods("DELETE")
	router.HandleFunc("/api/teams/import/{temp_id}", teamImportHandler.ImportTeam).Methods("POST")
	router.HandleFunc("/api/teams/import-all", teamImportHandler.ImportAllComplete).Methods("POST")

	// API - Competition (dodaj endpoint do pobierania info o scraperze)
	router.HandleFunc("/api/competition/current", func(w http.ResponseWriter, r *http.Request) {
		db := dbManager.GetDB()
		var competition models.Competition
		if err := db.First(&competition).Error; err != nil {
			http.Error(w, "Nie znaleziono competition", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "success",
			"competition": competition,
		})
	}).Methods("GET")

	// Start serwera
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("=================================")
	log.Printf("Serwer uruchomiony na: http://%s", addr)
	log.Printf("Panel WWW: http://localhost:%s", cfg.Server.Port)
	log.Printf("Zespoły: http://localhost:%s/teams", cfg.Server.Port)
	log.Printf("Socket.IO: ws://localhost:%s/socket.io/", cfg.Server.Port)
	if !needsSetup && dbConfig != nil {
		log.Printf("Aktualna baza danych: %s", dbManager.GetCurrentDatabaseName())
	} else {
		log.Printf("⚠️  WYMAGANA KONFIGURACJA - przejdź do /setup")
	}

	availableScrapers := scraperService.GetAvailableScrapers()
	log.Printf("Dostępne scrapery: %v", availableScrapers)
	// availableAlgorithms := tableService.GetAvailableAlgorithms()
	// log.Printf("Dostępne algorytmy tabel: %v", availableAlgorithms)

	log.Printf("=================================")

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Błąd uruchomienia serwera:", err)
	}
}

// checkSetupMiddleware - middleware sprawdzający czy aplikacja jest skonfigurowana
func checkSetupMiddleware(dbConfig *config.DatabaseConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Pomiń sprawdzanie dla ścieżek setup i static
			if strings.HasPrefix(r.URL.Path, "/setup") ||
				strings.HasPrefix(r.URL.Path, "/static") ||
				strings.HasPrefix(r.URL.Path, "/socket.io") {
				next.ServeHTTP(w, r)
				return
			}

			// Sprawdź czy istnieje konfiguracja
			if dbConfig == nil {
				if _, err := os.Stat("database_config.json"); os.IsNotExist(err) {
					http.Redirect(w, r, "/setup", http.StatusTemporaryRedirect)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// setupOBSEventHandlers - konfiguruje handlery eventów OBS
func setupOBSEventHandlers(obsClient *services.OBSClient) {
	// Event: Rozpoczęto/zatrzymano nagrywanie w OBS
	obsClient.OnEvent("RecordStateChanged", func(data map[string]interface{}) {
		outputActive, _ := data["outputActive"].(bool)
		if outputActive {
			log.Println("OBS Event: Nagrywanie rozpoczęte")
		} else {
			log.Println("OBS Event: Nagrywanie zatrzymane")
		}
	})

	// Event: Zmiana sceny
	obsClient.OnEvent("CurrentProgramSceneChanged", func(data map[string]interface{}) {
		sceneName, _ := data["sceneName"].(string)
		log.Printf("OBS Event: Zmieniono scenę na: %s", sceneName)
	})

	// Event: OBS uruchomiono/zamknięto
	obsClient.OnEvent("ExitStarted", func(data map[string]interface{}) {
		log.Println("OBS Event: Zamykanie aplikacji OBS")
	})
}
