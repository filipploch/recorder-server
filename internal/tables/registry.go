package tables

import (
	"fmt"
	"log"
	"sync"
)

// AlgorithmRegistry - rejestr algorytmów sortowania tabel
type AlgorithmRegistry struct {
	mu         sync.RWMutex
	algorithms map[string]TableOrderAlgorithm
}

var (
	algorithmRegistry *AlgorithmRegistry
	algorithmOnce     sync.Once
)

// GetAlgorithmRegistry - singleton registry algorytmów
func GetAlgorithmRegistry() *AlgorithmRegistry {
	algorithmOnce.Do(func() {
		algorithmRegistry = &AlgorithmRegistry{
			algorithms: make(map[string]TableOrderAlgorithm),
		}
		// Automatyczna rejestracja domyślnych algorytmów
		algorithmRegistry.registerDefaultAlgorithms()
	})
	return algorithmRegistry
}

// RegisterAlgorithm - rejestruje algorytm sortowania
func (r *AlgorithmRegistry) RegisterAlgorithm(name string, algorithm TableOrderAlgorithm) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.algorithms[name] = algorithm
	log.Printf("Table Algorithm Registry: Zarejestrowano algorytm '%s'", name)
}

// GetAlgorithm - pobiera algorytm po nazwie
func (r *AlgorithmRegistry) GetAlgorithm(name string) (TableOrderAlgorithm, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	algorithm, exists := r.algorithms[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrAlgorithmNotFound, name)
	}
	
	return algorithm, nil
}

// ListAlgorithms - zwraca listę dostępnych algorytmów
func (r *AlgorithmRegistry) ListAlgorithms() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	algorithms := make([]string, 0, len(r.algorithms))
	for name := range r.algorithms {
		algorithms = append(algorithms, name)
	}
	return algorithms
}

// HasAlgorithm - sprawdza czy algorytm istnieje
func (r *AlgorithmRegistry) HasAlgorithm(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.algorithms[name]
	return exists
}

// registerDefaultAlgorithms - rejestruje domyślne algorytmy
func (r *AlgorithmRegistry) registerDefaultAlgorithms() {
	// Przykładowa rejestracja - w przyszłości będą tu prawdziwe implementacje
	
	// r.RegisterAlgorithm("standard", &StandardAlgorithm{})
	// r.RegisterAlgorithm("head_to_head", &HeadToHeadAlgorithm{})
	// r.RegisterAlgorithm("goal_difference", &GoalDifferenceAlgorithm{})
	
	log.Println("Table Algorithm Registry: Inicjalizacja zakończona (brak domyślnych algorytmów)")
}
