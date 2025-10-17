package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"recorder-server/internal/models"
	"recorder-server/internal/services"
	"recorder-server/internal/state"
)

// CameraHandler - handler dla operacji na kamerach
type CameraHandler struct {
	appState      *state.AppState
	socketService *services.SocketIOService
}

// NewCameraHandler - tworzy nowy handler kamer
func NewCameraHandler(appState *state.AppState, socketService *services.SocketIOService) *CameraHandler {
	return &CameraHandler{
		appState:      appState,
		socketService: socketService,
	}
}

// StartRecording - rozpoczyna nagrywanie
func (h *CameraHandler) StartRecording(w http.ResponseWriter, r *http.Request) {
	var data models.StartRecordingData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	h.appState.StartRecording(data.ActiveCameras, data.InactiveCameras)
	h.socketService.BroadcastStartRecording(data)

	log.Printf("Rozpoczęto nagrywanie: aktywne=%v, nieaktywne=%v",
		data.ActiveCameras, data.InactiveCameras)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Status: "success"})
}

// StopRecording - zatrzymuje nagrywanie
func (h *CameraHandler) StopRecording(w http.ResponseWriter, r *http.Request) {
	var data models.StopRecordingData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	h.appState.StopRecording()
	h.socketService.BroadcastStopRecording(data)

	log.Printf("Zatrzymano nagrywanie dla kamer: %v", data.Cameras)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Status: "success"})
}

// GetRecordData - pobiera dane nagrywania
func (h *CameraHandler) GetRecordData(w http.ResponseWriter, r *http.Request) {
	var data models.GetRecordData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	h.socketService.BroadcastGetRecordData(data)

	log.Printf("Zapytanie o dane nagrywania: %+v", data)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Status: "request_sent"})
}

// GetStatus - pobiera status nagrywania
func (h *CameraHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.appState.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}