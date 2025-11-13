package scrapers

import (
	"context"
	"fmt"
	"log"
	"recorder-server/internal/models"
	"strings"
	"time"
	"unicode"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MZPNTeamScraper - scraper drużyn dla MZPN
type MZPNTeamScraper struct {
	*ExampleScraper
	baseURL string
}

// MZPNScraperWrapper - wrapper dla głównego scrapera MZPN
type MZPNScraperWrapper struct {
	*ExampleScraper
	baseURL     string
	teamScraper *MZPNTeamScraper
}

// NewMZPNScraperWrapper - tworzy wrapper scrapera dla MZPN
func NewMZPNScraperWrapper() *MZPNScraperWrapper {
	return &MZPNScraperWrapper{
		ExampleScraper: NewExampleScraper("MZPN"),
		baseURL:        "https://www.mzpnkrakow.pl",
		teamScraper:    NewMZPNTeamScraper(),
	}
}

// ScrapeTeams - implementacja dla MZPN (stara sygnatura - nieużywana)
func (s *MZPNScraperWrapper) ScrapeTeams(competitionID string) ([]models.Team, error) {
	return nil, fmt.Errorf("use ScrapeTeamsWithDB instead")
}

// ScrapeTeamsWithDB - nowa metoda używana przez handler
func (s *MZPNScraperWrapper) ScrapeTeamsWithDB(competitionID string, teamsURL string, db *gorm.DB) ([]models.TempTeam, error) {
	// Użyj dedykowanego team scrapera
	tempTeams, err := s.teamScraper.ScrapeTeams(competitionID, teamsURL, db)
	if err != nil {
		return nil, fmt.Errorf("MZPN team scraper error: %w", err)
	}

	// Wyświetl podsumowanie
	s.teamScraper.PrintSummary(tempTeams)

	return tempTeams, nil
}

// ScrapePlayers - implementacja interfejsu PlayerScraper
func (s *MZPNScraperWrapper) ScrapePlayers(teamID uint, externalTeamID string) ([]models.Player, error) {
	return nil, fmt.Errorf("not implemented: MZPN player scraper")
}

// ScrapeGames - implementacja interfejsu GameScraper
func (s *MZPNScraperWrapper) ScrapeGames(competitionID string, stageID uint) ([]models.Game, error) {
	return nil, fmt.Errorf("not implemented: MZPN game scraper")
}

// NewMZPNTeamScraper - tworzy scraper dla MZPN
func NewMZPNTeamScraper() *MZPNTeamScraper {
	return &MZPNTeamScraper{
		ExampleScraper: NewExampleScraper("MZPNTeams"),
		baseURL:        "https://www.mzpnkrakow.pl",
	}
}

// ScrapeTeams - scrapuje drużyny z podanego URL używając chromedp
func (s *MZPNTeamScraper) ScrapeTeams(competitionID string, teamsURL string, db *gorm.DB) ([]models.TempTeam, error) {
	log.Printf("MZPNScraper: Rozpoczynam scrapowanie drużyn z: %s", teamsURL)

	// Pobierz istniejące drużyny z bazy danych
	var existingTeams []models.Team
	db.Find(&existingTeams)
	existingForeignIDs := make(map[string]bool)
	for _, team := range existingTeams {
		if team.ForeignID != nil && *team.ForeignID != "" {
			existingForeignIDs[*team.ForeignID] = true
		}
	}

	// Pobierz istniejące drużyny z pliku tymczasowego
	manager := models.NewTempTeamManager(competitionID)
	tempCollection, _ := manager.Load()
	existingTempForeignIDs := make(map[string]bool)
	for _, team := range tempCollection.Teams {
		if team.ForeignID != nil && *team.ForeignID != "" {
			existingTempForeignIDs[*team.ForeignID] = true
		}
	}

	// Utwórz kontekst chromedp z timeoutem
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var teams []models.TempTeam

	// Uruchom chromedp task
	err := chromedp.Run(ctx,
		chromedp.Navigate(teamsURL),
		chromedp.WaitVisible(`table tbody`, chromedp.ByQuery),
	)

	if err != nil {
		return nil, fmt.Errorf("błąd chromedp: %w", err)
	}

	// Parsuj HTML używając chromedp do ekstrakcji danych
	teams, err = s.parseTeamsWithChromedp(ctx, teamsURL, existingForeignIDs, existingTempForeignIDs)
	if err != nil {
		return nil, fmt.Errorf("błąd parsowania drużyn: %w", err)
	}

	log.Printf("MZPNScraper: Znaleziono %d nowych drużyn (pominiętych duplikatów: %d)",
		len(teams),
		(len(existingForeignIDs) + len(existingTempForeignIDs)))
	return teams, nil
}

// parseTeamsWithChromedp - parsuje drużyny używając chromedp do ekstrakcji danych
func (s *MZPNTeamScraper) parseTeamsWithChromedp(ctx context.Context, teamsURL string,
	existingForeignIDs map[string]bool, existingTempForeignIDs map[string]bool) ([]models.TempTeam, error) {

	var teams []models.TempTeam

	// Pobierz liczbę wierszy TYLKO w pierwszej tabeli tbody
	var rowCount int
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('table tbody').querySelectorAll('tr').length`, &rowCount),
	)
	if err != nil {
		return nil, fmt.Errorf("błąd pobierania liczby wierszy: %w", err)
	}

	log.Printf("MZPNScraper: Znaleziono %d wierszy w pierwszej tabeli", rowCount)

	// Iteruj przez każdy wiersz TYLKO w pierwszej tabeli
	for i := 0; i < rowCount; i++ {
		var teamData map[string]interface{}

		jsCode := fmt.Sprintf(`
			(function() {
				// Pobierz TYLKO pierwszą tabelę
				const firstTable = document.querySelector('table');
				if (!firstTable) return null;
				
				const tbody = firstTable.querySelector('tbody');
				if (!tbody) return null;
				
				// Pobierz wiersz z PIERWSZEJ tabeli
				const row = tbody.querySelectorAll('tr')[%d];
				if (!row) return null;
				
				const nameCell = row.querySelector('td.text-start');
				if (!nameCell) return null;
				
				const teamNameRaw = nameCell.textContent.trim();
				
				return {
					name_raw: teamNameRaw
				};
			})()
		`, i)

		err := chromedp.Run(ctx,
			chromedp.Evaluate(jsCode, &teamData),
		)

		if err != nil {
			log.Printf("MZPNScraper: Błąd pobierania danych wiersza %d: %v", i, err)
			continue
		}

		if teamData == nil {
			log.Printf("MZPNScraper: Brak danych w wierszu %d", i)
			continue
		}

		nameRaw, _ := teamData["name_raw"].(string)

		if strings.TrimSpace(nameRaw) == "" {
			log.Printf("MZPNScraper: Pusta nazwa w wierszu %d, pomijam", i)
			continue
		}

		// Przekształć nazwę: pierwsza litera każdego wyrazu wielka, pozostałe małe
		name := toTitleCase(nameRaw)

		// Utwórz foreign_id: usuń wszystkie znaki białe z oryginalnej nazwy
		foreignID := removeWhitespace(nameRaw)

		// Pomiń jeśli już istnieje w bazie lub w pliku tymczasowym
		if foreignID != "" {
			if existingForeignIDs[foreignID] {
				log.Printf("MZPNScraper: Drużyna %s (foreign_id: %s) już istnieje w bazie, pomijam",
					name, foreignID)
				continue
			}
			if existingTempForeignIDs[foreignID] {
				log.Printf("MZPNScraper: Drużyna %s (foreign_id: %s) już istnieje w pliku tymczasowym, pomijam",
					name, foreignID)
				continue
			}
		}

		tempTeam := models.TempTeam{
			TempID:    uuid.New().String(),
			Name:      name,
			ShortName: nil,
			Name16:    nil,
			Logo:      nil,
			Link:      nil,
			ForeignID: stringPtr(foreignID),
			Source:    "mzpn_scraper",
			ScrapedAt: time.Now().Format(time.RFC3339),
			Notes:     "Zescrapowano automatycznie z mzpnkrakow.pl",
			Kits: map[string][]string{
				"1": {},
				"2": {},
				"3": {},
			},
		}

		teams = append(teams, tempTeam)
		log.Printf("MZPNScraper: [%d/%d] Dodano: %s (foreign_id: %s)",
			len(teams), rowCount, name, foreignID)
	}

	return teams, nil
}

// toTitleCase - przekształca tekst tak, że pierwsza litera każdego wyrazu jest wielka, pozostałe małe
func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(strings.ToLower(word))
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

// removeWhitespace - usuwa wszystkie znaki białe z tekstu
func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

// SaveTeamsToJSON - zapisuje drużyny do pliku JSON w katalogu tymczasowym
func (s *MZPNTeamScraper) SaveTeamsToJSON(competitionID string, teams []models.TempTeam) error {
	manager := models.NewTempTeamManager(competitionID)

	// Dodaj wszystkie drużyny masowo
	if err := manager.AddBulk(teams); err != nil {
		return fmt.Errorf("błąd zapisu drużyn do pliku: %w", err)
	}

	log.Printf("MZPNScraper: Zapisano %d drużyn do pliku tymczasowego", len(teams))
	return nil
}

// ClassifyTeams - klasyfikuje drużyny na kompletne i niekompletne
func (s *MZPNTeamScraper) ClassifyTeams(teams []models.TempTeam) (complete, incomplete []models.TempTeam) {
	for _, team := range teams {
		if team.IsComplete() {
			complete = append(complete, team)
		} else {
			incomplete = append(incomplete, team)
		}
	}
	return complete, incomplete
}

// PrintSummary - wyświetla podsumowanie scrapowania
func (s *MZPNTeamScraper) PrintSummary(teams []models.TempTeam) {
	complete, incomplete := s.ClassifyTeams(teams)

	log.Printf("========================================")
	log.Printf("PODSUMOWANIE SCRAPOWANIA")
	log.Printf("========================================")
	log.Printf("Łącznie drużyn:          %d", len(teams))
	log.Printf("Kompletne (gotowe):      %d", len(complete))
	log.Printf("Niekompletne (wymagają edycji): %d", len(incomplete))
	log.Printf("========================================")

	if len(incomplete) > 0 {
		log.Printf("Niekompletne drużyny wymagają uzupełnienia:")
		for _, team := range incomplete {
			missing := team.GetMissingFields()
			log.Printf("  - %s (brakuje: %v)", team.Name, missing)
		}
	}
}
