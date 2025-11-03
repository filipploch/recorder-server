package main

import (
	"fmt"
	"log"
	"os"
	"recorder-server/config"
	"recorder-server/internal/database"
	"recorder-server/internal/models"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Inicjalizuj manager
	dbManager := database.GetManager()
	if err := dbManager.Initialize(); err != nil {
		log.Fatal("Błąd inicjalizacji Database Manager:", err)
	}
	defer dbManager.Close()

	switch command {
	case "list":
		listDatabases(dbManager)
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Błąd: Brak nazwy bazy danych")
			fmt.Println("Użycie: db_helper create <nazwa_bazy>")
			os.Exit(1)
		}
		createDatabase(dbManager, os.Args[2])
	case "switch":
		if len(os.Args) < 3 {
			fmt.Println("Błąd: Brak nazwy bazy danych")
			fmt.Println("Użycie: db_helper switch <nazwa_bazy>")
			os.Exit(1)
		}
		switchDatabase(dbManager, os.Args[2])
	case "migrate":
		migrateDatabase(dbManager)
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Błąd: Brak nazwy bazy danych")
			fmt.Println("Użycie: db_helper delete <nazwa_bazy>")
			os.Exit(1)
		}
		deleteDatabase(dbManager, os.Args[2])
	case "current":
		showCurrent(dbManager)
	default:
		fmt.Printf("Nieznane polecenie: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("=== Database Helper ===")
	fmt.Println("Użycie: db_helper <polecenie> [argumenty]")
	fmt.Println("")
	fmt.Println("Dostępne polecenia:")
	fmt.Println("  list                    - Wyświetl listę wszystkich baz danych")
	fmt.Println("  current                 - Pokaż aktualną bazę danych")
	fmt.Println("  create <nazwa>          - Utwórz nową bazę danych i wykonaj migrację")
	fmt.Println("  switch <nazwa>          - Przełącz na inną bazę danych")
	fmt.Println("  migrate                 - Wykonaj migrację dla aktualnej bazy")
	fmt.Println("  delete <nazwa>          - Usuń bazę danych")
	fmt.Println("")
	fmt.Println("Przykłady:")
	fmt.Println("  db_helper list")
	fmt.Println("  db_helper create game_2025_01_15")
	fmt.Println("  db_helper switch game_2025_01_15")
	fmt.Println("  db_helper migrate")
	fmt.Println("  db_helper delete old_game")
}

func listDatabases(dbManager *database.Manager) {
	databases := dbManager.GetAvailableDatabases()
	current := dbManager.GetCurrentDatabaseName()

	fmt.Println("=== Lista baz danych ===")
	fmt.Printf("Aktualna baza: %s\n\n", current)
	fmt.Println("Dostępne bazy:")
	for i, db := range databases {
		marker := "  "
		if db == current {
			marker = "* "
		}
		fmt.Printf("%s%d. %s\n", marker, i+1, db)
	}
	fmt.Printf("\nRazem: %d baz danych\n", len(databases))
}

func createDatabase(dbManager *database.Manager, name string) {
	fmt.Printf("Tworzenie bazy danych: %s\n", name)
	
	if err := dbManager.CreateDatabase(name); err != nil {
		log.Fatal("Błąd tworzenia bazy:", err)
	}
	
	fmt.Println("✓ Baza utworzona")
	
	// Przełącz na nową bazę
	fmt.Println("Przełączanie na nową bazę...")
	if err := dbManager.SwitchDatabase(name); err != nil {
		log.Fatal("Błąd przełączania bazy:", err)
	}
	
	fmt.Println("✓ Przełączono na nową bazę")
	
	// Wykonaj migrację
	fmt.Println("Wykonywanie migracji...")
	if err := dbManager.AutoMigrate(models.GetAllModels()...); err != nil {
		log.Fatal("Błąd migracji:", err)
	}
	
	fmt.Printf("✓ Migracja zakończona pomyślnie\n")
	fmt.Printf("\nBaza danych '%s' została utworzona i jest gotowa do użycia!\n", name)
}

func switchDatabase(dbManager *database.Manager, name string) {
	fmt.Printf("Przełączanie na bazę: %s\n", name)
	
	if err := dbManager.SwitchDatabase(name); err != nil {
		log.Fatal("Błąd przełączania bazy:", err)
	}
	
	fmt.Println("✓ Przełączono na bazę:", name)
	
	// Wykonaj migrację dla pewności
	fmt.Println("Sprawdzanie/aktualizacja schematu...")
	if err := dbManager.AutoMigrate(models.GetAllModels()...); err != nil {
		log.Printf("Ostrzeżenie: Błąd migracji: %v", err)
	} else {
		fmt.Println("✓ Schemat bazy jest aktualny")
	}
}

func migrateDatabase(dbManager *database.Manager) {
	current := dbManager.GetCurrentDatabaseName()
	fmt.Printf("Wykonywanie migracji dla bazy: %s\n", current)
	
	if err := dbManager.AutoMigrate(models.GetAllModels()...); err != nil {
		log.Fatal("Błąd migracji:", err)
	}
	
	fmt.Println("✓ Migracja zakończona pomyślnie")
	
	// Pokaż listę tabel
	db := dbManager.GetDB()
	var tables []string
	
	// Pobierz listę tabel (dla SQLite)
	db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables)
	
	fmt.Println("\nUtwórzone tabele:")
	for i, table := range tables {
		fmt.Printf("  %d. %s\n", i+1, table)
	}
}

func deleteDatabase(dbManager *database.Manager, name string) {
	current := dbManager.GetCurrentDatabaseName()
	
	if name == current {
		fmt.Printf("BŁĄD: Nie można usunąć aktualnie używanej bazy (%s)\n", name)
		fmt.Println("Najpierw przełącz się na inną bazę używając: db_helper switch <nazwa>")
		os.Exit(1)
	}
	
	fmt.Printf("Czy na pewno chcesz usunąć bazę danych '%s'? (tak/nie): ", name)
	var confirm string
	fmt.Scanln(&confirm)
	
	if confirm != "tak" {
		fmt.Println("Anulowano usuwanie bazy danych")
		return
	}
	
	fmt.Printf("Usuwanie bazy danych: %s\n", name)
	
	if err := dbManager.DeleteDatabase(name); err != nil {
		log.Fatal("Błąd usuwania bazy:", err)
	}
	
	fmt.Printf("✓ Baza danych '%s' została usunięta\n", name)
}

func showCurrent(dbManager *database.Manager) {
	current := dbManager.GetCurrentDatabaseName()
	
	fmt.Println("=== Aktualna baza danych ===")
	fmt.Printf("Nazwa: %s\n", current)
	
	// Pokaż statystyki
	db := dbManager.GetDB()
	
	fmt.Println("\nTabele:")
	
	models := []struct {
		Name  string
		Model interface{}
	}{
		{"Teams", &models.Team{}},
		{"Players", &models.Player{}},
		{"Coaches", &models.Coach{}},
		{"Referees", &models.Referee{}},
		{"Fields", &models.Field{}},
		{"Competitions", &models.Competition{}},
		{"Games", &models.Game{}},
		{"GameParts", &models.GamePart{}},
		{"TVStaff", &models.TVStaff{}},
		{"Events", &models.Event{}},
	}
	
	for _, m := range models {
		var count int64
		db.Model(m.Model).Count(&count)
		fmt.Printf("  %-15s %d rekordów\n", m.Name+":", count)
	}
}