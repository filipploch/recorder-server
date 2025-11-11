package scrapers

import (
	"context"
	//	"encoding/json"
	"fmt"
	"log"
	"recorder-server/internal/models"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NalffutsalTeamScraper - scraper drużyn dla NALF Futsal
type NalffutsalTeamScraper struct {
	*ExampleScraper
	baseURL string
}

// NewNalffutsalTeamScraper - tworzy scraper dla NALF Futsal
func NewNalffutsalTeamScraper() *NalffutsalTeamScraper {
	return &NalffutsalTeamScraper{
		ExampleScraper: NewExampleScraper("NalffutsalTeams"),
		baseURL:        "https://nalffutsal.pl",
	}
}

// ScrapeTeams - scrapuje drużyny z podanego URL używając chromedp
func (s *NalffutsalTeamScraper) ScrapeTeams(competitionID string, teamsURL string, db *gorm.DB) ([]models.TempTeam, error) {
	log.Printf("NalffutsalScraper: Rozpoczynam scrapowanie drużyn z: %s", teamsURL)

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

	log.Printf("NalffutsalScraper: Znaleziono %d nowych drużyn (pominiętych duplikatów: %d)",
		len(teams),
		(len(existingForeignIDs) + len(existingTempForeignIDs)))
	return teams, nil
}

// parseTeamsWithChromedp - parsuje drużyny używając chromedp do ekstrakcji danych
func (s *NalffutsalTeamScraper) parseTeamsWithChromedp(ctx context.Context, teamsURL string,
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

	log.Printf("NalffutsalScraper: Znaleziono %d wierszy w pierwszej tabeli", rowCount)

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
				
				const nameCell = row.querySelector('td.data-name');
				if (!nameCell) return null;
				
				const link = nameCell.querySelector('a');
				if (!link) return null;
				
				const teamName = link.textContent.trim();
				
				const href = link.getAttribute('href');
				let foreignId = '';
				if (href) {
					const match = href.match(/sp_team=([^&]+)/);
					if (match) {
						foreignId = match[1];
					}
				}
				
				return {
					name: teamName,
					link: href || '',
					foreign_id: foreignId
				};
			})()
		`, i)

		err := chromedp.Run(ctx,
			chromedp.Evaluate(jsCode, &teamData),
		)

		if err != nil {
			log.Printf("NalffutsalScraper: Błąd pobierania danych wiersza %d: %v", i, err)
			continue
		}

		if teamData == nil {
			log.Printf("NalffutsalScraper: Brak danych w wierszu %d", i)
			continue
		}

		name, _ := teamData["name"].(string)
		link, _ := teamData["link"].(string)
		foreignID, _ := teamData["foreign_id"].(string)

		if strings.TrimSpace(name) == "" {
			log.Printf("NalffutsalScraper: Pusta nazwa w wierszu %d, pomijam", i)
			continue
		}

		// Pomiń jeśli już istnieje w bazie lub w pliku tymczasowym
		if foreignID != "" {
			if existingForeignIDs[foreignID] {
				log.Printf("NalffutsalScraper: Drużyna %s (foreign_id: %s) już istnieje w bazie, pomijam",
					name, foreignID)
				continue
			}
			if existingTempForeignIDs[foreignID] {
				log.Printf("NalffutsalScraper: Drużyna %s (foreign_id: %s) już istnieje w pliku tymczasowym, pomijam",
					name, foreignID)
				continue
			}
		}

		tempTeam := models.TempTeam{
			TempID:    uuid.New().String(),
			Name:      strings.TrimSpace(name),
			ShortName: nil,
			Name16:    nil,
			Logo:      nil,
			Link:      stringPtr(link),
			ForeignID: stringPtr(foreignID),
			Source:    "nalffutsal_scraper",
			ScrapedAt: time.Now().Format(time.RFC3339),
			Notes:     "Zescrapowano automatycznie z nalffutsal.pl",
		}

		teams = append(teams, tempTeam)
		log.Printf("NalffutsalScraper: [%d/%d] Dodano: %s (foreign_id: %s)",
			len(teams), rowCount, name, foreignID)
	}

	return teams, nil
}

// stringPtr - helper do tworzenia wskaźników na string
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// SaveTeamsToJSON - zapisuje drużyny do pliku JSON w katalogu tymczasowym
func (s *NalffutsalTeamScraper) SaveTeamsToJSON(competitionID string, teams []models.TempTeam) error {
	manager := models.NewTempTeamManager(competitionID)

	// Dodaj wszystkie drużyny masowo
	if err := manager.AddBulk(teams); err != nil {
		return fmt.Errorf("błąd zapisu drużyn do pliku: %w", err)
	}

	log.Printf("NalffutsalScraper: Zapisano %d drużyn do pliku tymczasowego", len(teams))
	return nil
}

// ClassifyTeams - klasyfikuje drużyny na kompletne i niekompletne
func (s *NalffutsalTeamScraper) ClassifyTeams(teams []models.TempTeam) (complete, incomplete []models.TempTeam) {
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
func (s *NalffutsalTeamScraper) PrintSummary(teams []models.TempTeam) {
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
