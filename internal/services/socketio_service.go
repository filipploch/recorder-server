package services

import (
	"log"
	"recorder-server/internal/models"
	"recorder-server/internal/state"

	socketio "github.com/googollee/go-socket.io"
)

// SocketIOService - serwis Socket.IO
type SocketIOService struct {
	server   *socketio.Server
	appState *state.AppState
}

// NewSocketIOService - tworzy nowy serwis Socket.IO
func NewSocketIOService(appState *state.AppState) *SocketIOService {
	server := socketio.NewServer(nil)

	service := &SocketIOService{
		server:   server,
		appState: appState,
	}

	service.setupHandlers()
	return service
}

// setupHandlers - konfiguruje handlery Socket.IO
func (s *SocketIOService) setupHandlers() {
	s.server.OnConnect("/", func(conn socketio.Conn) error {
		log.Printf("Socket.IO: Nowe połączenie: %s", conn.ID())
		conn.Join("room1") // Dołącz do pokoju
		return nil
	})

	s.server.OnDisconnect("/", func(conn socketio.Conn, reason string) {
		log.Printf("Socket.IO: Rozłączenie: %s, powód: %s", conn.ID(), reason)
	})

	s.server.OnEvent("/", "get_status", func(conn socketio.Conn) {
		status := s.appState.GetStatus()
		conn.Emit("status_response", status)
		log.Printf("Socket.IO: Wysłano status do klienta: %+v", status)
	})

	s.server.OnError("/", func(conn socketio.Conn, e error) {
		log.Printf("Socket.IO: Błąd: %v", e)
	})
}

// GetServer - zwraca serwer Socket.IO
func (s *SocketIOService) GetServer() *socketio.Server {
	return s.server
}

// BroadcastStartRecording - rozgłasza sygnał start_recording
func (s *SocketIOService) BroadcastStartRecording(data models.StartRecordingData) {
	s.server.BroadcastToRoom("/", "room1", "start_recording", data)
	log.Printf("Socket.IO: Broadcast start_recording: %+v", data)
}

// BroadcastStopRecording - rozgłasza sygnał stop_recording
func (s *SocketIOService) BroadcastStopRecording(data models.StopRecordingData) {
	s.server.BroadcastToRoom("/", "room1", "stop_recording", data)
	log.Printf("Socket.IO: Broadcast stop_recording: %+v", data)
}

// BroadcastGetRecordData - rozgłasza zapytanie o dane nagrywania
func (s *SocketIOService) BroadcastGetRecordData(data models.GetRecordData) {
	s.server.BroadcastToRoom("/", "room1", "get_record_data", data)
	log.Printf("Socket.IO: Broadcast get_record_data: %+v", data)
}

// BroadcastTimerUpdate - rozgłasza aktualizację stopera
func (s *SocketIOService) BroadcastTimerUpdate(data interface{}) {
	// Użyj BroadcastToNamespace zamiast BroadcastToRoom
	s.server.BroadcastToNamespace("/", "timer_update", data)
	log.Printf("Socket.IO: Broadcast timer_update: %+v", data)
}