package timer

import (
	"log"
	"sync"
	"time"
)

// Engine - silnik stopera
type Engine struct {
	config            Config
	formatter         *Formatter
	mu                sync.RWMutex
	running           bool
	startTime         time.Time
	pausedElapsed     int64 // Czas który upłynął przed pauzą
	ticker            *time.Ticker
	stopChan          chan bool
	broadcastCallback func(UpdateMessage)
	lastBroadcastTime time.Time
}

// NewEngine - tworzy nowy silnik stopera
func NewEngine(config Config) *Engine {
	return &Engine{
		config:    config,
		formatter: NewFormatter(config.BroadcastPrecision),
		stopChan:  make(chan bool),
	}
}

// SetBroadcastCallback - ustawia funkcję callback dla broadcastu
func (e *Engine) SetBroadcastCallback(callback func(UpdateMessage)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.broadcastCallback = callback
}

// Start - rozpoczyna stoper
func (e *Engine) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}

	e.running = true
	e.startTime = time.Now()
	e.pausedElapsed = 0
	
	measurementInterval := PrecisionToDuration(e.config.MeasurementPrecision)
	e.ticker = time.NewTicker(measurementInterval)
	
	// Utwórz nowy kanał stop
	e.stopChan = make(chan bool, 1)
	
	e.mu.Unlock()

	log.Printf("Timer: Uruchomiono stoper (kierunek: %s, max: %v, pomiar: %v, broadcast: %v)",
		DirectionToString(e.config.Direction),
		e.config.MaxDuration,
		measurementInterval,
		PrecisionToDuration(e.config.BroadcastPrecision))

	// Wyślij początkowy stan
	e.broadcast()

	go e.run()
}

// Stop - zatrzymuje stoper
func (e *Engine) Stop() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}

	e.running = false
	if e.ticker != nil {
		e.ticker.Stop()
	}
	e.mu.Unlock()

	e.stopChan <- true
	log.Println("Timer: Zatrzymano stoper")

	// Wyślij końcowy stan
	e.broadcast()
}

// Pause - pauzuje stoper
func (e *Engine) Pause() {
	e.mu.Lock()
	
	if !e.running {
		e.mu.Unlock()
		log.Println("Timer: Stoper już zapauzowany")
		return
	}

	// Oblicz rzeczywisty upłynięty czas
	realElapsed := e.pausedElapsed + time.Since(e.startTime).Milliseconds()
	
	// Dla direction UP zapisz upłynięty czas
	// Dla direction DOWN zapisz pozostały czas, ale ograniczony do max
	if e.config.Direction == DirectionDown && e.config.MaxDuration != nil {
		remaining := *e.config.MaxDuration - realElapsed
		if remaining < 0 {
			remaining = 0
		}
		e.pausedElapsed = remaining
	} else if e.config.Direction == DirectionUp && e.config.MaxDuration != nil {
		// Dla UP, jeśli przekroczono max, zapisz max (chyba że continue mode)
		maxMs := *e.config.MaxDuration
		if realElapsed > maxMs && e.config.StopBehavior == StopBehaviorAuto {
			e.pausedElapsed = maxMs
		} else {
			e.pausedElapsed = realElapsed
		}
	} else {
		// Bez limitu lub direction UP bez ograniczenia
		e.pausedElapsed = realElapsed
	}
	
	e.running = false
	
	if e.ticker != nil {
		e.ticker.Stop()
		e.ticker = nil
	}
	e.mu.Unlock()

	// Wyślij sygnał stop do goroutine
	select {
	case e.stopChan <- true:
	default:
	}

	log.Printf("Timer: Zapauzowano stoper (pausedElapsed: %dms, direction: %s)", 
		e.pausedElapsed, DirectionToString(e.config.Direction))
	
	// Broadcast aktualnego stanu PO pauzie
	e.broadcast()
}

// Resume - wznawia stoper
func (e *Engine) Resume() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		log.Println("Timer: Stoper już działa, ignoruję Resume")
		return
	}

	e.running = true
	e.startTime = time.Now()
	
	measurementInterval := PrecisionToDuration(e.config.MeasurementPrecision)
	e.ticker = time.NewTicker(measurementInterval)
	
	// Utwórz nowy kanał stop
	e.stopChan = make(chan bool, 1)
	
	e.mu.Unlock()

	log.Printf("Timer: Wznowiono stoper (pausedElapsed: %dms)", e.pausedElapsed)
	
	// Wyślij aktualny stan
	e.broadcast()
	
	go e.run()
}

// Reset - resetuje stoper
func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()

	wasRunning := e.running
	
	// Zatrzymaj jeśli działa
	if e.running {
		e.running = false
		if e.ticker != nil {
			e.ticker.Stop()
			e.ticker = nil
		}
	}

	// Wyczyść wszystkie wartości
	e.pausedElapsed = 0
	e.startTime = time.Now()

	log.Println("Timer: Zresetowano stoper")
	
	// Wyślij sygnał stop jeśli działał
	if wasRunning {
		select {
		case e.stopChan <- true:
		default:
		}
	}
	
	// Broadcast zresetowanego stanu
	e.broadcast()
}

// run - główna pętla stopera
func (e *Engine) run() {
	broadcastInterval := PrecisionToDuration(e.config.BroadcastPrecision)

	for {
		select {
		case <-e.ticker.C:
			e.mu.RLock()
			if !e.running {
				e.mu.RUnlock()
				return
			}

			elapsed := e.pausedElapsed + time.Since(e.startTime).Milliseconds()
			
			// Sprawdź czy osiągnięto maksymalny czas
			shouldStop := false
			if e.config.MaxDuration != nil {
				maxMs := *e.config.MaxDuration
				
				if e.config.Direction == DirectionUp {
					// Direction UP: sprawdź czy elapsed >= max
					if elapsed >= maxMs && e.config.StopBehavior == StopBehaviorAuto {
						shouldStop = true
					}
				} else {
					// Direction DOWN: sprawdź czy pozostały czas <= 0
					remaining := maxMs - elapsed
					if remaining <= 0 && e.config.StopBehavior == StopBehaviorAuto {
						shouldStop = true
					}
				}
			}
			
			e.mu.RUnlock()
			
			// Jeśli osiągnięto limit, zapauzuj
			if shouldStop {
				log.Println("Timer: Osiągnięto maksymalny czas, automatyczne zatrzymanie")
				e.Pause()
				return
			}

			// Broadcast jeśli minął odpowiedni czas
			if time.Since(e.lastBroadcastTime) >= broadcastInterval {
				e.broadcast()
			}

		case <-e.stopChan:
			return
		}
	}
}

// calculateElapsed - oblicza wartość do wyświetlenia
func (e *Engine) calculateElapsed() int64 {
	// Jeśli stoper nie działa, zwróć zapisaną wartość
	if !e.running {
		return e.pausedElapsed
	}
	
	// Oblicz rzeczywisty upłynięty czas od startu
	realElapsed := e.pausedElapsed + time.Since(e.startTime).Milliseconds()
	
	// Dla direction DOWN zwróć pozostały czas
	if e.config.Direction == DirectionDown && e.config.MaxDuration != nil {
		remaining := *e.config.MaxDuration - realElapsed
		if remaining < 0 {
			remaining = 0
		}
		return remaining
	}
	
	// Dla direction UP zwróć upłynięty czas
	return realElapsed
}

// broadcast - wysyła aktualny stan przez callback
func (e *Engine) broadcast() {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.broadcastCallback == nil {
		log.Println("Timer: Brak callback dla broadcast")
		return
	}

	// calculateElapsed zwraca wartość do wyświetlenia
	// DOWN: pozostały czas
	// UP: upłynięty czas
	displayTime := e.calculateElapsed()
	
	// Sprawdź overflow tylko dla continue mode
	var overflowMs int64
	var isOverflow bool
	
	if e.config.MaxDuration != nil && e.config.StopBehavior == StopBehaviorContinue {
		if e.config.Direction == DirectionUp {
			maxMs := *e.config.MaxDuration
			realElapsed := e.pausedElapsed
			if e.running {
				realElapsed += time.Since(e.startTime).Milliseconds()
			}
			
			if realElapsed > maxMs {
				isOverflow = true
				overflowMs = realElapsed - maxMs
				displayTime = maxMs
			}
		} else if e.config.Direction == DirectionDown {
			if displayTime == 0 {
				realElapsed := e.pausedElapsed
				if e.running {
					realElapsed = e.pausedElapsed + time.Since(e.startTime).Milliseconds()
				}
				
				remaining := *e.config.MaxDuration - realElapsed
				if remaining < 0 {
					isOverflow = true
					overflowMs = -remaining
					displayTime = 0
				}
			}
		}
	}

	var formattedTime string
	if isOverflow {
		formattedTime = e.formatter.FormatWithOverflow(displayTime, overflowMs)
	} else {
		formattedTime = e.formatter.Format(displayTime)
	}

	e.lastBroadcastTime = time.Now()

	msg := UpdateMessage{
		ElapsedMs:     displayTime,
		FormattedTime: formattedTime,
		IsOverflow:    isOverflow,
		OverflowMs:    overflowMs,
	}

	log.Printf("Timer: Broadcasting - %s (display: %dms, running: %v, dir: %s)", 
		formattedTime, displayTime, e.running, DirectionToString(e.config.Direction))

	go e.broadcastCallback(msg)
}

// GetState - pobiera aktualny stan stopera
func (e *Engine) GetState() State {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// calculateElapsed zwraca wartość do wyświetlenia
	displayTime := e.calculateElapsed()
	
	// Sprawdź overflow tylko dla continue mode
	var overflowMs int64
	var isOverflow bool
	
	if e.config.MaxDuration != nil && e.config.StopBehavior == StopBehaviorContinue {
		if e.config.Direction == DirectionUp {
			maxMs := *e.config.MaxDuration
			realElapsed := e.pausedElapsed
			if e.running {
				realElapsed += time.Since(e.startTime).Milliseconds()
			}
			
			if realElapsed > maxMs {
				isOverflow = true
				overflowMs = realElapsed - maxMs
				displayTime = maxMs
			}
		} else if e.config.Direction == DirectionDown {
			if displayTime == 0 {
				realElapsed := e.pausedElapsed
				if e.running {
					realElapsed = e.pausedElapsed + time.Since(e.startTime).Milliseconds()
				}
				
				remaining := *e.config.MaxDuration - realElapsed
				if remaining < 0 {
					isOverflow = true
					overflowMs = -remaining
					displayTime = 0
				}
			}
		}
	}

	var formattedTime string
	if isOverflow {
		formattedTime = e.formatter.FormatWithOverflow(displayTime, overflowMs)
	} else {
		formattedTime = e.formatter.Format(displayTime)
	}

	return State{
		Running:        e.running,
		ElapsedMs:      displayTime,
		MaxDurationMs:  e.config.MaxDuration,
		Direction:      DirectionToString(e.config.Direction),
		OverflowMs:     overflowMs,
		FormattedTime:  formattedTime,
		IsOverflow:     isOverflow,
	}
}

// IsRunning - sprawdza czy stoper działa
func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}