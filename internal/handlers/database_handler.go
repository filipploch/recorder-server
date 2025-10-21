package handlers

import (
	"encoding/json"
	"net/http"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
)

// DatabaseHandler - handler dla operacji na bazach danych
type DatabaseHandler struct {
	manager *database.Manager
}

// NewDatabaseHandler - tworzy nowy handler bazy danych
func NewDatabaseHandler(manager *database.Manager) *DatabaseHandler {
	return &DatabaseHandler{
		manager: manager,
	}
}

// GetCurrent - zwraca informacje o aktualnej bazie
func (h *DatabaseHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	currentDB := h.manager.GetCurrentDatabaseName()
	
	response := map[string]interface{}{
		"current_database": currentDB,
		"status":           "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAvailable - zwraca listę dostępnych baz
func (h *DatabaseHandler) GetAvailable(w http.ResponseWriter, r *http.Request) {
	databases := h.manager.GetAvailableDatabases()
	currentDB := h.manager.GetCurrentDatabaseName()

	response := map[string]interface{}{
		"databases":        databases,
		"current_database": currentDB,
		"count":            len(databases),
		"status":           "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SwitchDatabase - przełącza na inną bazę i wykonuje migrację
func (h *DatabaseHandler) SwitchDatabase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DatabaseName string `json:"database_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	if req.DatabaseName == "" {
		http.Error(w, "Nazwa bazy jest wymagana", http.StatusBadRequest)
		return
	}

	// Przełącz bazę
	if err := h.manager.SwitchDatabase(req.DatabaseName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	// Wykonaj migrację dla nowej bazy
	if err := h.manager.AutoMigrate(models.GetAllModels()...); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  "Przełączono bazę, ale migracja nie powiodła się: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":           "success",
		"current_database": req.DatabaseName,
		"message":          "Przełączono na bazę: " + req.DatabaseName + " i wykonano migrację",
	})
}

// CreateDatabase - tworzy nową bazę danych i wykonuje migrację
func (h *DatabaseHandler) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DatabaseName string `json:"database_name"`
		SwitchTo     bool   `json:"switch_to"` // Czy automatycznie przełączyć na nową bazę
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	if req.DatabaseName == "" {
		http.Error(w, "Nazwa bazy jest wymagana", http.StatusBadRequest)
		return
	}

	// Utwórz bazę
	if err := h.manager.CreateDatabase(req.DatabaseName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	// Jeśli switch_to = true, przełącz na nową bazę
	if req.SwitchTo {
		if err := h.manager.SwitchDatabase(req.DatabaseName); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.APIResponse{
				Status: "error",
				Error:  "Utworzono bazę, ale nie można było przełączyć: " + err.Error(),
			})
			return
		}
	}

	// Wykonaj migrację dla nowej bazy
	if err := h.manager.AutoMigrate(models.GetAllModels()...); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  "Utworzono bazę, ale migracja nie powiodła się: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"status":        "success",
		"database_name": req.DatabaseName,
		"message":       "Utworzono bazę: " + req.DatabaseName + " i wykonano migrację",
	}

	if req.SwitchTo {
		response["current_database"] = req.DatabaseName
		response["message"] = "Utworzono bazę i przełączono na: " + req.DatabaseName
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// DeleteDatabase - usuwa bazę danych
func (h *DatabaseHandler) DeleteDatabase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DatabaseName string `json:"database_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	if req.DatabaseName == "" {
		http.Error(w, "Nazwa bazy jest wymagana", http.StatusBadRequest)
		return
	}

	if err := h.manager.DeleteDatabase(req.DatabaseName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Usunięto bazę: " + req.DatabaseName,
	})
}
