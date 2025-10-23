package services

import (
	"recorder-server/internal/database"
	"recorder-server/internal/models"
	"sync"
)

// SessionService - serwis zarządzający aktywną sesją
type SessionService struct {
	mu      sync.RWMutex
	manager *database.Manager
	cache   *models.ActiveSession
}

// NewSessionService - tworzy nowy serwis sesji
func NewSessionService(manager *database.Manager) *SessionService {
	service := &SessionService{
		manager: manager,
	}
	service.LoadSession()
	return service
}

// LoadSession - wczytuje aktywną sesję z bazy do cache
func (s *SessionService) LoadSession() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	db := s.manager.GetDB()
	if db == nil {
		return nil
	}

	var session models.ActiveSession
	if err := db.Preload("Game").Preload("GamePart").First(&session).Error; err != nil {
		// Jeśli nie ma sesji, utwórz pustą
		session = models.ActiveSession{
			GameID:     nil,
			GamePartID: nil,
		}
		db.Create(&session)
	}

	s.cache = &session
	return nil
}

// GetActiveSession - zwraca aktywną sesję
func (s *SessionService) GetActiveSession() *models.ActiveSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.cache == nil {
		s.LoadSession()
	}
	
	return s.cache
}

// SetActiveGame - ustawia aktywny mecz
func (s *SessionService) SetActiveGame(gameID *uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	db := s.manager.GetDB()
	if db == nil {
		return nil
	}

	if s.cache == nil {
		s.LoadSession()
	}

	s.cache.GameID = gameID
	s.cache.GamePartID = nil // Reset części meczu

	if err := db.Save(s.cache).Error; err != nil {
		return err
	}

	// Przeładuj z relacjami
	db.Preload("Game").Preload("GamePart").First(s.cache)
	return nil
}

// SetActiveGamePart - ustawia aktywną część meczu
func (s *SessionService) SetActiveGamePart(gamePartID *uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	db := s.manager.GetDB()
	if db == nil {
		return nil
	}

	if s.cache == nil {
		s.LoadSession()
	}

	s.cache.GamePartID = gamePartID

	if err := db.Save(s.cache).Error; err != nil {
		return err
	}

	// Przeładuj z relacjami
	db.Preload("Game").Preload("GamePart").First(s.cache)
	return nil
}

// ClearSession - czyści aktywną sesję
func (s *SessionService) ClearSession() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	db := s.manager.GetDB()
	if db == nil {
		return nil
	}

	if s.cache == nil {
		s.LoadSession()
	}

	s.cache.GameID = nil
	s.cache.GamePartID = nil

	if err := db.Save(s.cache).Error; err != nil {
		return err
	}

	return nil
}

// GetActiveGameID - zwraca ID aktywnego meczu
func (s *SessionService) GetActiveGameID() *uint {
	session := s.GetActiveSession()
	if session != nil {
		return session.GameID
	}
	return nil
}

// GetActiveGamePartID - zwraca ID aktywnej części meczu
func (s *SessionService) GetActiveGamePartID() *uint {
	session := s.GetActiveSession()
	if session != nil {
		return session.GamePartID
	}
	return nil
}
