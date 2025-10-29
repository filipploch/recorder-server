package handlers

import (
	"encoding/json"
	"net/http"
	"recorder-server/internal/models"
	"recorder-server/internal/services"
	"strconv"
)

// TableHandler - handler dla operacji na tabelach
type TableHandler struct {
	tableService *services.TableService
}

// NewTableHandler - tworzy nowy handler tabel
func NewTableHandler(tableService *services.TableService) *TableHandler {
	return &TableHandler{
		tableService: tableService,
	}
}

// CalculateTableForGroup - endpoint do obliczania tabeli dla grupy
// POST /api/tables/group/:id/calculate
// lub GET /api/tables/group?group_id=1
func (h *TableHandler) CalculateTableForGroup(w http.ResponseWriter, r *http.Request) {
	// Pobierz group_id z query
	groupIDStr := r.URL.Query().Get("group_id")
	if groupIDStr == "" {
		http.Error(w, "Brak group_id", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Nieprawidłowe group_id", http.StatusBadRequest)
		return
	}

	// Oblicz tabelę
	table, err := h.tableService.CalculateTableForGroup(uint(groupID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"table":  table,
	})
}

// CalculateTableForStage - endpoint do obliczania tabel dla stage
// GET /api/tables/stage?stage_id=1
func (h *TableHandler) CalculateTableForStage(w http.ResponseWriter, r *http.Request) {
	stageIDStr := r.URL.Query().Get("stage_id")
	if stageIDStr == "" {
		http.Error(w, "Brak stage_id", http.StatusBadRequest)
		return
	}

	stageID, err := strconv.ParseUint(stageIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Nieprawidłowe stage_id", http.StatusBadRequest)
		return
	}

	// Oblicz tabele
	tables, err := h.tableService.CalculateTableForStage(uint(stageID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"tables": tables,
		"count":  len(tables),
	})
}

// CalculateTableForCompetition - endpoint do obliczania tabel dla competition
// GET /api/tables/competition?competition_id=1
func (h *TableHandler) CalculateTableForCompetition(w http.ResponseWriter, r *http.Request) {
	competitionIDStr := r.URL.Query().Get("competition_id")
	if competitionIDStr == "" {
		http.Error(w, "Brak competition_id", http.StatusBadRequest)
		return
	}

	competitionID, err := strconv.ParseUint(competitionIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Nieprawidłowe competition_id", http.StatusBadRequest)
		return
	}

	// Oblicz tabele
	tables, err := h.tableService.CalculateTableForCompetition(uint(competitionID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"tables": tables,
		"count":  len(tables),
	})
}

// GetAvailableAlgorithms - endpoint zwracający listę dostępnych algorytmów
// GET /api/tables/algorithms
func (h *TableHandler) GetAvailableAlgorithms(w http.ResponseWriter, r *http.Request) {
	algorithms := h.tableService.GetAvailableAlgorithms()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"algorithms": algorithms,
		"count":      len(algorithms),
	})
}

// GetCompetitionAlgorithmInfo - endpoint zwracający info o algorytmie dla competition
// GET /api/tables/competition/:id/algorithm
func (h *TableHandler) GetCompetitionAlgorithmInfo(w http.ResponseWriter, r *http.Request) {
	competitionIDStr := r.URL.Query().Get("competition_id")
	if competitionIDStr == "" {
		http.Error(w, "Brak competition_id", http.StatusBadRequest)
		return
	}

	competitionID, err := strconv.ParseUint(competitionIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Nieprawidłowe competition_id", http.StatusBadRequest)
		return
	}

	// TODO: Pobierz competition z bazy i zwróć info o algorytmie
	_ = competitionID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":               "success",
		"competition_id":       competitionID,
		"table_order_algorithm": "standard", // TODO: Pobierz z bazy
	})
}

// CompareTeams - endpoint do porównania dwóch drużyn (debug)
// GET /api/tables/compare?group_id=1&team1_id=1&team2_id=2
func (h *TableHandler) CompareTeams(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.URL.Query().Get("group_id")
	team1IDStr := r.URL.Query().Get("team1_id")
	team2IDStr := r.URL.Query().Get("team2_id")

	if groupIDStr == "" || team1IDStr == "" || team2IDStr == "" {
		http.Error(w, "Brak wymaganych parametrów", http.StatusBadRequest)
		return
	}

	groupID, _ := strconv.ParseUint(groupIDStr, 10, 32)
	team1ID, _ := strconv.ParseUint(team1IDStr, 10, 32)
	team2ID, _ := strconv.ParseUint(team2IDStr, 10, 32)

	result, err := h.tableService.CompareTeamsInGroup(uint(groupID), uint(team1ID), uint(team2ID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	var resultText string
	switch result {
	case -1:
		resultText = "Team 1 wyżej"
	case 0:
		resultText = "Równe"
	case 1:
		resultText = "Team 2 wyżej"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"result":      result,
		"result_text": resultText,
	})
}