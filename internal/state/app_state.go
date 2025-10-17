package state

import (
	"recorder-server/internal/models"
	"sync"
)

// AppState - globalny stan aplikacji
type AppState struct {
	mu              sync.RWMutex
	isRecording     bool
	activeCameras   []string
	inactiveCameras []string
	allCameras      []string
}

// NewAppState - tworzy nowy stan aplikacji
func NewAppState(allCameras []string) *AppState {
	return &AppState{
		allCameras:      allCameras,
		isRecording:     false,
		activeCameras:   []string{},
		inactiveCameras: allCameras,
	}
}

// GetStatus - pobiera status nagrywania
func (s *AppState) GetStatus() models.RecordStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return models.RecordStatus{
		RecordStatus:    s.isRecording,
		ActiveCameras:   append([]string{}, s.activeCameras...),
		InactiveCameras: append([]string{}, s.inactiveCameras...),
	}
}

// StartRecording - rozpoczyna nagrywanie
func (s *AppState) StartRecording(active, inactive []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isRecording = true
	s.activeCameras = append([]string{}, active...)
	s.inactiveCameras = append([]string{}, inactive...)
}

// StopRecording - zatrzymuje nagrywanie
func (s *AppState) StopRecording() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isRecording = false
	s.activeCameras = []string{}
	s.inactiveCameras = s.allCameras
}

// IsRecording - sprawdza czy trwa nagrywanie
func (s *AppState) IsRecording() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRecording
}

// GetActiveCameras - pobiera aktywne kamery
func (s *AppState) GetActiveCameras() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string{}, s.activeCameras...)
}

// GetAllCameras - pobiera wszystkie kamery
func (s *AppState) GetAllCameras() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string{}, s.allCameras...)
}