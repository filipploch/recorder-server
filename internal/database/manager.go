package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"recorder-server/config"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	
)

// Manager - zarządza wieloma bazami danych
type Manager struct {
	mu              sync.RWMutex
	currentDB       *gorm.DB
	currentName     string
	dbConfig        *config.DatabaseConfig
	databases       map[string]*gorm.DB // Cache otwartych połączeń
}

var (
	instance *Manager
	once     sync.Once
)

// GetManager - zwraca singleton instancję managera
func GetManager() *Manager {
	once.Do(func() {
		instance = &Manager{
			databases: make(map[string]*gorm.DB),
		}
	})
	return instance
}

// Initialize - inicjalizuje manager z konfiguracją
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Wczytaj konfigurację
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		return fmt.Errorf("błąd wczytywania konfiguracji bazy: %w", err)
	}
	m.dbConfig = dbConfig

	// Utwórz katalog na bazy danych jeśli nie istnieje
	if err := os.MkdirAll(m.dbConfig.DatabasesPath, 0755); err != nil {
		return fmt.Errorf("błąd tworzenia katalogu baz danych: %w", err)
	}

	// Połącz z aktualną bazą
	if err := m.connectToDatabase(m.dbConfig.CurrentDatabase); err != nil {
		return fmt.Errorf("błąd połączenia z bazą %s: %w", m.dbConfig.CurrentDatabase, err)
	}

	log.Printf("Database Manager: Zainicjalizowano (current: %s)", m.currentName)
	return nil
}

// connectToDatabase - łączy z konkretną bazą danych
func (m *Manager) connectToDatabase(dbName string) error {
	// Sprawdź czy już mamy otwarte połączenie
	if db, exists := m.databases[dbName]; exists {
		m.currentDB = db
		m.currentName = dbName
		log.Printf("Database: Użyto cached połączenia: %s", dbName)
		return nil
	}

	// Ścieżka do pliku bazy
	dbPath := filepath.Join(m.dbConfig.DatabasesPath, dbName+".db")

	// Otwórz połączenie z GORM
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Zmień na logger.Info dla debugowania
	})
	if err != nil {
		return fmt.Errorf("błąd otwierania bazy %s: %w", dbName, err)
	}

	// Zapisz w cache
	m.databases[dbName] = db
	m.currentDB = db
	m.currentName = dbName

	log.Printf("Database: Połączono z bazą: %s (%s)", dbName, dbPath)
	return nil
}

// SwitchDatabase - przełącza na inną bazę danych
func (m *Manager) SwitchDatabase(dbName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Połącz z nową bazą
	if err := m.connectToDatabase(dbName); err != nil {
		return err
	}

	// Aktualizuj konfigurację
	m.dbConfig.CurrentDatabase = dbName

	// Dodaj do listy dostępnych jeśli jeszcze nie ma
	found := false
	for _, name := range m.dbConfig.AvailableDatabases {
		if name == dbName {
			found = true
			break
		}
	}
	if !found {
		m.dbConfig.AvailableDatabases = append(m.dbConfig.AvailableDatabases, dbName)
	}

	// Zapisz konfigurację
	if err := config.SaveDatabaseConfig(m.dbConfig); err != nil {
		log.Printf("Database: Ostrzeżenie - nie udało się zapisać konfiguracji: %v", err)
	}

	log.Printf("Database: Przełączono na bazę: %s", dbName)
	return nil
}

// GetDB - zwraca aktualną instancję bazy danych
func (m *Manager) GetDB() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentDB
}

// GetCurrentDatabaseName - zwraca nazwę aktualnej bazy
func (m *Manager) GetCurrentDatabaseName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentName
}

// GetAvailableDatabases - zwraca listę dostępnych baz
func (m *Manager) GetAvailableDatabases() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Zwróć kopię aby uniknąć race conditions
	databases := make([]string, len(m.dbConfig.AvailableDatabases))
	copy(databases, m.dbConfig.AvailableDatabases)
	return databases
}

// CreateDatabase - tworzy nową bazę danych
func (m *Manager) CreateDatabase(dbName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Sprawdź czy baza już istnieje
	dbPath := filepath.Join(m.dbConfig.DatabasesPath, dbName+".db")
	if _, err := os.Stat(dbPath); err == nil {
		return fmt.Errorf("baza danych %s już istnieje", dbName)
	}

	// Utwórz nową bazę (otwarcie połączenia utworzy plik)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("błąd tworzenia bazy %s: %w", dbName, err)
	}

	// Zapisz w cache
	m.databases[dbName] = db

	// Dodaj do listy dostępnych
	m.dbConfig.AvailableDatabases = append(m.dbConfig.AvailableDatabases, dbName)

	// Zapisz konfigurację
	if err := config.SaveDatabaseConfig(m.dbConfig); err != nil {
		log.Printf("Database: Ostrzeżenie - nie udało się zapisać konfiguracji: %v", err)
	}

	log.Printf("Database: Utworzono nową bazę: %s", dbName)
	return nil
}

// DeleteDatabase - usuwa bazę danych
func (m *Manager) DeleteDatabase(dbName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Nie pozwól usunąć aktualnej bazy
	if dbName == m.currentName {
		return fmt.Errorf("nie można usunąć aktualnie używanej bazy")
	}

	// Zamknij połączenie jeśli istnieje
	if db, exists := m.databases[dbName]; exists {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		delete(m.databases, dbName)
	}

	// Usuń plik
	dbPath := filepath.Join(m.dbConfig.DatabasesPath, dbName+".db")
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("błąd usuwania pliku bazy: %w", err)
	}

	// Usuń z listy dostępnych
	newList := []string{}
	for _, name := range m.dbConfig.AvailableDatabases {
		if name != dbName {
			newList = append(newList, name)
		}
	}
	m.dbConfig.AvailableDatabases = newList

	// Zapisz konfigurację
	if err := config.SaveDatabaseConfig(m.dbConfig); err != nil {
		log.Printf("Database: Ostrzeżenie - nie udało się zapisać konfiguracji: %v", err)
	}

	log.Printf("Database: Usunięto bazę: %s", dbName)
	return nil
}

// Close - zamyka wszystkie połączenia z bazami
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, db := range m.databases {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Database: Błąd pobierania *sql.DB dla %s: %v", name, err)
			continue
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("Database: Błąd zamykania połączenia %s: %v", name, err)
		} else {
			log.Printf("Database: Zamknięto połączenie: %s", name)
		}
	}

	m.databases = make(map[string]*gorm.DB)
	m.currentDB = nil
	m.currentName = ""

	return nil
}

// AutoMigrate - wykonuje auto-migrację dla wszystkich modeli
func (m *Manager) AutoMigrate(models ...interface{}) error {
	m.mu.RLock()
	db := m.currentDB
	m.mu.RUnlock()

	if db == nil {
		return fmt.Errorf("brak aktywnego połączenia z bazą danych")
	}

	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("błąd migracji: %w", err)
	}

	log.Printf("Database: Wykonano migrację dla %d modeli w bazie %s", len(models), m.currentName)
	return nil
}