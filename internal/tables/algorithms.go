package tables

import (
	"fmt"
	"recorder-server/internal/models"
	"sort"
	"time"
)

// StandardAlgorithm - standardowy algorytm sortowania (punkty > różnica bramek > bramki zdobyte)
type StandardAlgorithm struct{}

func (a *StandardAlgorithm) GetName() string {
	return "standard"
}

func (a *StandardAlgorithm) CalculateTable(groupID uint, teams []models.Team, games []models.Game) (*Table, error) {
	// TODO: Implementacja obliczania tabeli
	// 1. Dla każdej drużyny zlicz: mecze, wygrane, remisy, przegrane, bramki
	// 2. Oblicz punkty (zwykle 3 za wygraną, 1 za remis)
	// 3. Sortuj według punktów, różnicy bramek, bramek zdobytych
	
	standings := make([]TeamStanding, 0, len(teams))
	
	// Przykładowa struktura - w rzeczywistości trzeba obliczyć na podstawie games
	for i, team := range teams {
		standings = append(standings, TeamStanding{
			Position:       i + 1,
			Team:           &team,
			TeamID:         team.ID,
			Played:         0,
			Won:            0,
			Drawn:          0,
			Lost:           0,
			GoalsFor:       0,
			GoalsAgainst:   0,
			GoalDifference: 0,
			Points:         0,
		})
	}
	
	// Sortuj
	sort.Slice(standings, func(i, j int) bool {
		return a.CompareTeams(&standings[i], &standings[j], nil) < 0
	})
	
	// Aktualizuj pozycje
	for i := range standings {
		standings[i].Position = i + 1
	}
	
	return &Table{
		GroupID:   groupID,
		Standings: standings,
		UpdatedAt: time.Now().Format(time.RFC3339),
		Algorithm: a.GetName(),
	}, nil
}

func (a *StandardAlgorithm) CompareTeams(s1, s2 *TeamStanding, headToHeadGames []models.Game) int {
	// Najpierw punkty
	if s1.Points != s2.Points {
		if s1.Points > s2.Points {
			return -1 // s1 wyżej
		}
		return 1 // s2 wyżej
	}
	
	// Potem różnica bramek
	if s1.GoalDifference != s2.GoalDifference {
		if s1.GoalDifference > s2.GoalDifference {
			return -1
		}
		return 1
	}
	
	// Potem bramki zdobyte
	if s1.GoalsFor != s2.GoalsFor {
		if s1.GoalsFor > s2.GoalsFor {
			return -1
		}
		return 1
	}
	
	// Równe
	return 0
}

// HeadToHeadAlgorithm - algorytm z uwzględnieniem meczów bezpośrednich
type HeadToHeadAlgorithm struct{}

func (a *HeadToHeadAlgorithm) GetName() string {
	return "head_to_head"
}

func (a *HeadToHeadAlgorithm) CalculateTable(groupID uint, teams []models.Team, games []models.Game) (*Table, error) {
	// TODO: Podobnie jak StandardAlgorithm, ale z uwzględnieniem meczów bezpośrednich
	
	return nil, fmt.Errorf("head_to_head algorithm not implemented yet")
}

func (a *HeadToHeadAlgorithm) CompareTeams(s1, s2 *TeamStanding, headToHeadGames []models.Game) int {
	// Najpierw punkty
	if s1.Points != s2.Points {
		if s1.Points > s2.Points {
			return -1
		}
		return 1
	}
	
	// Potem mecze bezpośrednie
	if len(headToHeadGames) > 0 {
		// TODO: Analiza meczów bezpośrednich
		// - punkty w meczach bezpośrednich
		// - różnica bramek w meczach bezpośrednich
		// - bramki zdobyte na wyjeździe w meczach bezpośrednich
	}
	
	// Potem różnica bramek ogólna
	if s1.GoalDifference != s2.GoalDifference {
		if s1.GoalDifference > s2.GoalDifference {
			return -1
		}
		return 1
	}
	
	return 0
}

// GoalDifferenceFirstAlgorithm - algorytm priorytetyzujący różnicę bramek
type GoalDifferenceFirstAlgorithm struct{}

func (a *GoalDifferenceFirstAlgorithm) GetName() string {
	return "goal_difference_first"
}

func (a *GoalDifferenceFirstAlgorithm) CalculateTable(groupID uint, teams []models.Team, games []models.Game) (*Table, error) {
	// TODO: Implementacja
	return nil, fmt.Errorf("goal_difference_first algorithm not implemented yet")
}

func (a *GoalDifferenceFirstAlgorithm) CompareTeams(s1, s2 *TeamStanding, headToHeadGames []models.Game) int {
	// Najpierw różnica bramek (!)
	if s1.GoalDifference != s2.GoalDifference {
		if s1.GoalDifference > s2.GoalDifference {
			return -1
		}
		return 1
	}
	
	// Potem punkty
	if s1.Points != s2.Points {
		if s1.Points > s2.Points {
			return -1
		}
		return 1
	}
	
	return 0
}

// RegisterDefaultAlgorithms - rejestruje domyślne algorytmy
// Ta funkcja powinna być wywołana podczas inicjalizacji aplikacji
func RegisterDefaultAlgorithms() {
	registry := GetAlgorithmRegistry()
	
	registry.RegisterAlgorithm("standard", &StandardAlgorithm{})
	registry.RegisterAlgorithm("head_to_head", &HeadToHeadAlgorithm{})
	registry.RegisterAlgorithm("goal_difference_first", &GoalDifferenceFirstAlgorithm{})
}
