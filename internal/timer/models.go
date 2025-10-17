package timer

import "time"

// Precision - dokładność pomiaru czasu
type Precision int

const (
	PrecisionSecond      Precision = iota // 1s
	PrecisionDecisecond                   // 0.1s
	PrecisionCentisecond                  // 0.01s
	PrecisionMillisecond                  // 0.001s
)

// Direction - kierunek liczenia czasu
type Direction int

const (
	DirectionUp   Direction = iota // Liczenie w górę (0 -> max)
	DirectionDown                   // Liczenie w dół (max -> 0)
)

// StopBehavior - zachowanie po osiągnięciu maksymalnego czasu
type StopBehavior int

const (
	StopBehaviorAuto     StopBehavior = iota // Automatyczne zatrzymanie
	StopBehaviorContinue                     // Kontynuuj z dodatkowym czasem
)

// Config - konfiguracja stopera
type Config struct {
	MeasurementPrecision Precision    // Dokładność pomiaru (jak często aktualizować wewnętrzny licznik)
	BroadcastPrecision   Precision    // Dokładność broadcast (jak często wysyłać przez Socket.IO)
	Direction            Direction    // Kierunek liczenia
	MaxDuration          *int64       // Maksymalny czas w ms (nil = nieokreślony)
	StopBehavior         StopBehavior // Zachowanie po osiągnięciu max
}

// State - stan stopera
type State struct {
	Running        bool   `json:"running"`         // Czy stoper działa
	ElapsedMs      int64  `json:"elapsed_ms"`      // Upłynięty czas w ms
	MaxDurationMs  *int64 `json:"max_duration_ms"` // Maksymalny czas w ms
	Direction      string `json:"direction"`       // "up" lub "down"
	OverflowMs     int64  `json:"overflow_ms"`     // Czas przekroczenia (dla continue mode)
	FormattedTime  string `json:"formatted_time"`  // Sformatowany czas do wyświetlenia
	IsOverflow     bool   `json:"is_overflow"`     // Czy przekroczono max
}

// StartRequest - żądanie rozpoczęcia stopera
type StartRequest struct {
	MeasurementPrecision string `json:"measurement_precision"` // "s", "ds", "cs", "ms"
	BroadcastPrecision   string `json:"broadcast_precision"`   // "s", "ds", "cs", "ms"
	Direction            string `json:"direction"`             // "up" lub "down"
	MaxDuration          *int64 `json:"max_duration"`          // w sekundach (nil = nieokreślony)
	StopBehavior         string `json:"stop_behavior"`         // "auto" lub "continue"
}

// UpdateMessage - wiadomość aktualizacji czasu
type UpdateMessage struct {
	ElapsedMs     int64  `json:"elapsed_ms"`
	FormattedTime string `json:"formatted_time"`
	IsOverflow    bool   `json:"is_overflow"`
	OverflowMs    int64  `json:"overflow_ms"`
}

// Helper functions

// PrecisionToDuration - konwertuje Precision na time.Duration
func PrecisionToDuration(p Precision) time.Duration {
	switch p {
	case PrecisionSecond:
		return time.Second
	case PrecisionDecisecond:
		return 100 * time.Millisecond
	case PrecisionCentisecond:
		return 10 * time.Millisecond
	case PrecisionMillisecond:
		return time.Millisecond
	default:
		return time.Second
	}
}

// PrecisionFromString - konwertuje string na Precision
func PrecisionFromString(s string) Precision {
	switch s {
	case "s":
		return PrecisionSecond
	case "ds":
		return PrecisionDecisecond
	case "cs":
		return PrecisionCentisecond
	case "ms":
		return PrecisionMillisecond
	default:
		return PrecisionSecond
	}
}

// DirectionToString - konwertuje Direction na string
func DirectionToString(d Direction) string {
	if d == DirectionUp {
		return "up"
	}
	return "down"
}

// DirectionFromString - konwertuje string na Direction
func DirectionFromString(s string) Direction {
	if s == "down" {
		return DirectionDown
	}
	return DirectionUp
}

// StopBehaviorFromString - konwertuje string na StopBehavior
func StopBehaviorFromString(s string) StopBehavior {
	if s == "continue" {
		return StopBehaviorContinue
	}
	return StopBehaviorAuto
}