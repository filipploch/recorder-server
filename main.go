package main

import (
	"log"
	"net/http"
	"recorder-server/config"
	"recorder-server/internal/database"
	"recorder-server/internal/handlers"
	"recorder-server/internal/models"
	"recorder-server/internal/services"
	"recorder-server/internal/state"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("=== Recorder Server ===")
	log.Println("Inicjalizacja aplikacji...")

	// Załaduj konfigurację
	cfg := config.LoadConfig()
	log.Printf("Konfiguracja załadowana: Port=%s, OBS URL=%s", cfg.Server.Port, cfg.OBS.URL)

	// Inicjalizacja stanu aplikacji
	appState := state.NewAppState(cfg.Recording.AllCameras)
	log.Println("Stan aplikacji zainicjalizowany")

	// Inicjalizacja serwisu Socket.IO
	socketService := services.NewSocketIOService(appState)
	log.Println("Socket.IO serwis zainicjalizowany")

	// Inicjalizacja serwisu stopera
	timerService := services.NewTimerService(socketService)
	log.Println("Timer serwis zainicjalizowany")

	// Inicjalizacja Database Manager
	dbManager := database.GetManager()
	if err := dbManager.Initialize(); err != nil {
		log.Fatal("Błąd inicjalizacji Database Manager:", err)
	}
	log.Println("Database Manager zainicjalizowany")
	
	// Wykonaj migrację modeli dla aktualnej bazy
	log.Println("Wykonywanie migracji bazy danych...")
	if err := dbManager.AutoMigrate(models.GetAllModels()...); err != nil {
		log.Printf("OSTRZEŻENIE: Błąd migracji bazy danych: %v", err)
	} else {
		log.Printf("Migracja zakończona pomyślnie dla bazy: %s", dbManager.GetCurrentDatabaseName())
	}
	
	// Zamknij połączenia z bazami przy wyjściu
	defer dbManager.Close()

	// Inicjalizacja klienta OBS WebSocket
	obsClient := services.NewOBSClient(cfg.OBS.URL, cfg.OBS.Password)
	go obsClient.Connect()
	setupOBSEventHandlers(obsClient)
	log.Println("OBS WebSocket klient zainicjalizowany")

	// Inicjalizacja handlerów
	pageHandler := handlers.NewPageHandler()
	cameraHandler := handlers.NewCameraHandler(appState, socketService)
	obsHandler := handlers.NewOBSHandler(obsClient)
	timerHandler := handlers.NewTimerHandler(timerService)
	databaseHandler := handlers.NewDatabaseHandler(dbManager)
	log.Println("Handlery HTTP zainicjalizowane")

	// Router
	router := mux.NewRouter()

	// Socket.IO - MUSI BYĆ PRZED INNYMI ROUTAMI
	router.Handle("/socket.io/", socketService.GetServer())

	// Strony WWW
	router.HandleFunc("/", pageHandler.Index).Methods("GET")

	// Pliki statyczne
	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))),
	)

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

	// Start serwera
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("=================================")
	log.Printf("Serwer uruchomiony na: http://%s", addr)
	log.Printf("Panel WWW: http://localhost:%s", cfg.Server.Port)
	log.Printf("Socket.IO: ws://localhost:%s/socket.io/", cfg.Server.Port)
	log.Printf("Aktualna baza danych: %s", dbManager.GetCurrentDatabaseName())
	log.Printf("=================================")

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Błąd uruchomienia serwera:", err)
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