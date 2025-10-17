package services

import (
	"log"
	"recorder-server/internal/timer"
	"sync"
)

// TimerService - serwis zarządzający stoperem
type TimerService struct {
	engine            *timer.Engine
	mu                sync.RWMutex
	socketService     *SocketIOService
	updateCallback    func(timer.UpdateMessage)
}

// NewTimerService - tworzy nowy serwis stopera
func NewTimerService(socketService *SocketIOService) *TimerService {
	service := &TimerService{
		socketService: socketService,
	}
	return service
}

// Initialize - inicjalizuje stoper z konfiguracją
func (s *TimerService) Initialize(config timer.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Zatrzymaj poprzedni silnik jeśli istnieje
	if s.engine != nil && s.engine.IsRunning() {
		s.engine.Stop()
	}

	// Utwórz nowy silnik
	s.engine = timer.NewEngine(config)

	// Ustaw callback dla broadcastu
	s.engine.SetBroadcastCallback(func(msg timer.UpdateMessage) {
		if s.updateCallback != nil {
			s.updateCallback(msg)
		}
		// Broadcast przez Socket.IO
		s.socketService.BroadcastTimerUpdate(msg)
	})

	log.Printf("TimerService: Zainicjalizowano z konfiguracją: %+v", config)
}

// Start - rozpoczyna lub wznawia stoper
func (s *TimerService) Start(req timer.StartRequest) error {
	s.mu.Lock()
	
	// Jeśli silnik już istnieje
	if s.engine != nil {
		// Sprawdź czy jest zapauzowany (nie działa)
		if !s.engine.IsRunning() {
			s.mu.Unlock()
			log.Println("TimerService: Wznawianie zapauzowanego stopera")
			s.engine.Resume()
			return nil
		}
		
		// Jeśli już działa, nie rób nic
		s.mu.Unlock()
		log.Println("TimerService: Stoper już działa")
		return nil
	}
	
	// Silnik nie istnieje, utwórz nowy
	s.mu.Unlock()
	
	log.Println("TimerService: Tworzenie nowego stopera")
	
	// Konwertuj request na config
	config := timer.Config{
		MeasurementPrecision: timer.PrecisionFromString(req.MeasurementPrecision),
		BroadcastPrecision:   timer.PrecisionFromString(req.BroadcastPrecision),
		Direction:            timer.DirectionFromString(req.Direction),
		StopBehavior:         timer.StopBehaviorFromString(req.StopBehavior),
	}

	// Konwertuj max duration z sekund na milisekundy
	if req.MaxDuration != nil {
		maxMs := *req.MaxDuration * 1000
		config.MaxDuration = &maxMs
	}

	// Inicjalizuj
	s.Initialize(config)
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.engine == nil {
		log.Println("TimerService: Błąd - brak silnika stopera")
		return nil
	}

	// Uruchom
	s.engine.Start()
	return nil
}

// Pause - pauzuje stoper
func (s *TimerService) Pause() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.engine != nil {
		s.engine.Pause()
	}
}

// Reset - resetuje stoper do stanu początkowego
func (s *TimerService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Zatrzymaj i usuń obecny silnik
	if s.engine != nil {
		if s.engine.IsRunning() {
			s.engine.Stop()
		}
		s.engine = nil
	}
	
	log.Println("TimerService: Reset stopera")
}

// GetState - pobiera aktualny stan stopera
func (s *TimerService) GetState() timer.State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.engine == nil {
		return timer.State{
			Running:       false,
			ElapsedMs:     0,
			Direction:     "up",
			FormattedTime: "00:00",
		}
	}

	return s.engine.GetState()
}

// SetUpdateCallback - ustawia callback dla aktualizacji
func (s *TimerService) SetUpdateCallback(callback func(timer.UpdateMessage)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateCallback = callback
}