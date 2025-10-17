package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"recorder-server/internal/models"
	"recorder-server/internal/services"
	"recorder-server/internal/timer"
)

// TimerHandler - handler dla operacji stopera
type TimerHandler struct {
	timerService *services.TimerService
}

// NewTimerHandler - tworzy nowy handler stopera
func NewTimerHandler(timerService *services.TimerService) *TimerHandler {
	return &TimerHandler{
		timerService: timerService,
	}
}

// Start - rozpoczyna lub wznawia stoper
func (h *TimerHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req timer.StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	// Ustaw domyślne wartości jeśli nie podano
	if req.MeasurementPrecision == "" {
		req.MeasurementPrecision = "ms" // 0.001s
	}
	if req.BroadcastPrecision == "" {
		req.BroadcastPrecision = "ds" // 0.1s
	}
	if req.Direction == "" {
		req.Direction = "up"
	}
	if req.StopBehavior == "" {
		req.StopBehavior = "auto"
	}

	err := h.timerService.Start(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	log.Printf("Timer: Uruchomiono/wznowiono stoper z konfiguracją: %+v", req)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Status: "success"})
}

// Pause - pauzuje stoper
func (h *TimerHandler) Pause(w http.ResponseWriter, r *http.Request) {
	h.timerService.Pause()
	
	log.Println("Timer: Zapauzowano stoper")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Status: "success"})
}

// Reset - resetuje stoper
func (h *TimerHandler) Reset(w http.ResponseWriter, r *http.Request) {
	h.timerService.Reset()
	
	log.Println("Timer: Zresetowano stoper")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Status: "success"})
}

// GetState - pobiera aktualny stan stopera
func (h *TimerHandler) GetState(w http.ResponseWriter, r *http.Request) {
	state := h.timerService.GetState()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}