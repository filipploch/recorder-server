package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"recorder-server/internal/models"
	"recorder-server/internal/services"
)

// OBSHandler - handler dla operacji OBS
type OBSHandler struct {
	obsClient *services.OBSClient
}

// NewOBSHandler - tworzy nowy handler OBS
func NewOBSHandler(obsClient *services.OBSClient) *OBSHandler {
	return &OBSHandler{
		obsClient: obsClient,
	}
}

// StartRecording - rozpoczyna nagrywanie w OBS
func (h *OBSHandler) StartRecording(w http.ResponseWriter, r *http.Request) {
	if !h.obsClient.IsConnected() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  "OBS not connected",
		})
		return
	}

	err := h.obsClient.StartRecording()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	log.Println("OBS: Rozpoczęto nagrywanie przez API")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Status: "success"})
}

// StopRecording - zatrzymuje nagrywanie w OBS
func (h *OBSHandler) StopRecording(w http.ResponseWriter, r *http.Request) {
	if !h.obsClient.IsConnected() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  "OBS not connected",
		})
		return
	}

	err := h.obsClient.StopRecording()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	log.Println("OBS: Zatrzymano nagrywanie przez API")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Status: "success"})
}

// GetStatus - pobiera status OBS
func (h *OBSHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	connected := h.obsClient.IsConnected()

	response := models.OBSStatusResponse{
		Connected: connected,
		Recording: false,
	}

	if connected {
		recording, err := h.obsClient.GetRecordingStatus()
		if err == nil {
			response.Recording = recording
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetScenes - pobiera listę scen OBS
func (h *OBSHandler) GetScenes(w http.ResponseWriter, r *http.Request) {
	if !h.obsClient.IsConnected() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(models.OBSScenesResponse{
			Status: "error",
			Error:  "OBS not connected",
		})
		return
	}

	scenes, err := h.obsClient.GetSceneList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.OBSScenesResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.OBSScenesResponse{
		Status: "success",
		Scenes: scenes,
	})
}

// SetScene - ustawia aktywną scenę
func (h *OBSHandler) SetScene(w http.ResponseWriter, r *http.Request) {
	if !h.obsClient.IsConnected() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  "OBS not connected",
		})
		return
	}

	var data models.SetSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Błąd dekodowania JSON", http.StatusBadRequest)
		return
	}

	err := h.obsClient.SetCurrentScene(data.SceneName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	log.Printf("OBS: Zmieniono scenę na: %s", data.SceneName)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Status: "success"})
}