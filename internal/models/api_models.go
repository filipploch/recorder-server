package models

// Modele używane do komunikacji API (nie są modelami bazy danych)

// RecordStatus - status nagrywania
type RecordStatus struct {
	RecordStatus    bool     `json:"record_status"`
	ActiveCameras   []string `json:"active_cameras"`
	InactiveCameras []string `json:"inactive_cameras"`
}

// StartRecordingData - dane dla rozpoczęcia nagrywania
type StartRecordingData struct {
	ActiveCameras   []string `json:"active_cameras"`
	InactiveCameras []string `json:"inactive_cameras"`
}

// StopRecordingData - dane dla zatrzymania nagrywania
type StopRecordingData struct {
	Cameras []string `json:"cameras"`
}

// GetRecordData - zapytanie o dane nagrywania
type GetRecordData struct {
	EventID       string   `json:"event_id"`
	ActiveCameras []string `json:"active_cameras"`
}

// RecordDataResponse - odpowiedź z danymi nagrywania
type RecordDataResponse struct {
	EventID         string `json:"event_id"`
	CameraName      string `json:"camera_name"`
	FileName        string `json:"file_name"`
	RecordStartTime string `json:"record_start_time"`
	CurrentTime     string `json:"current_time"`
}

// OBSStatusResponse - odpowiedź ze statusem OBS
type OBSStatusResponse struct {
	Connected bool `json:"connected"`
	Recording bool `json:"recording"`
}

// OBSScenesResponse - odpowiedź z listą scen OBS
type OBSScenesResponse struct {
	Status string   `json:"status"`
	Scenes []string `json:"scenes"`
	Error  string   `json:"error,omitempty"`
}

// SetSceneRequest - żądanie zmiany sceny
type SetSceneRequest struct {
	SceneName string `json:"scene_name"`
}

// APIResponse - generyczna odpowiedź API
type APIResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
