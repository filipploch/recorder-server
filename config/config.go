package config

// Config - konfiguracja aplikacji
type Config struct {
	Server    ServerConfig
	OBS       OBSConfig
	SocketIO  SocketIOConfig
	Recording RecordingConfig
}

// ServerConfig - konfiguracja serwera HTTP
type ServerConfig struct {
	Host string
	Port string
}

// OBSConfig - konfiguracja OBS WebSocket
type OBSConfig struct {
	URL      string
	Password string
}

// SocketIOConfig - konfiguracja Socket.IO
type SocketIOConfig struct {
	Enabled bool
}

// RecordingConfig - konfiguracja nagrywania
type RecordingConfig struct {
	AllCameras []string
}

// LoadConfig - ładuje konfigurację
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: "8080",
		},
		OBS: OBSConfig{
			URL:      "ws://localhost:4445",
			Password: "", // Ustaw hasło jeśli OBS wymaga
		},
		SocketIO: SocketIOConfig{
			Enabled: true,
		},
		Recording: RecordingConfig{
			AllCameras: []string{
				"camera_main",
				"camera_center",
				"camera_left",
				"camera_right",
			},
		},
	}
}