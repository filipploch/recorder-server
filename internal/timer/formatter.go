package timer

import "fmt"

// Formatter - formatuje czas do wyświetlenia
type Formatter struct {
	precision Precision
}

// NewFormatter - tworzy nowy formatter
func NewFormatter(precision Precision) *Formatter {
	return &Formatter{
		precision: precision,
	}
}

// Format - formatuje czas w milisekundach do stringa
func (f *Formatter) Format(ms int64) string {
	if ms < 0 {
		ms = 0
	}

	hours := ms / (1000 * 60 * 60)
	ms %= 1000 * 60 * 60

	minutes := ms / (1000 * 60)
	ms %= 1000 * 60

	seconds := ms / 1000
	ms %= 1000

	deciseconds := ms / 100
	centiseconds := ms / 10
	milliseconds := ms

	switch f.precision {
	case PrecisionSecond:
		if hours > 0 {
			return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		}
		return fmt.Sprintf("%02d:%02d", minutes, seconds)

	case PrecisionDecisecond:
		if hours > 0 {
			return fmt.Sprintf("%02d:%02d:%02d.%01d", hours, minutes, seconds, deciseconds)
		}
		return fmt.Sprintf("%02d:%02d.%01d", minutes, seconds, deciseconds)

	case PrecisionCentisecond:
		if hours > 0 {
			return fmt.Sprintf("%02d:%02d:%02d.%02d", hours, minutes, seconds, centiseconds)
		}
		return fmt.Sprintf("%02d:%02d.%02d", minutes, seconds, centiseconds)

	case PrecisionMillisecond:
		if hours > 0 {
			return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
		}
		return fmt.Sprintf("%02d:%02d.%03d", minutes, seconds, milliseconds)

	default:
		return fmt.Sprintf("%02d:%02d", minutes, seconds)
	}
}

// FormatWithOverflow - formatuje czas z informacją o przekroczeniu
func (f *Formatter) FormatWithOverflow(ms int64, overflowMs int64) string {
	baseFormat := f.Format(ms)
	if overflowMs > 0 {
		overflowFormat := f.Format(overflowMs)
		return fmt.Sprintf("%s (+%s)", baseFormat, overflowFormat)
	}
	return baseFormat
}

// ParseDuration - parsuje string czasu do milisekund
// Format: "HH:MM:SS" lub "MM:SS" lub "SS"
func ParseDuration(s string) (int64, error) {
	var hours, minutes, seconds int64
	
	// Próba parsowania HH:MM:SS
	n, err := fmt.Sscanf(s, "%d:%d:%d", &hours, &minutes, &seconds)
	if err == nil && n == 3 {
		return (hours*3600 + minutes*60 + seconds) * 1000, nil
	}

	// Próba parsowania MM:SS
	n, err = fmt.Sscanf(s, "%d:%d", &minutes, &seconds)
	if err == nil && n == 2 {
		return (minutes*60 + seconds) * 1000, nil
	}

	// Próba parsowania SS
	n, err = fmt.Sscanf(s, "%d", &seconds)
	if err == nil && n == 1 {
		return seconds * 1000, nil
	}

	return 0, fmt.Errorf("nieprawidłowy format czasu: %s", s)
}