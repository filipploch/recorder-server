@echo off
echo ========================================
echo  Recorder Server - Setup Script
echo ========================================
echo.

REM Utwórz główną strukturę katalogów
echo Tworzenie struktury katalogów...

mkdir config 2>nul
mkdir internal\models 2>nul
mkdir internal\handlers 2>nul
mkdir internal\services 2>nul
mkdir internal\state 2>nul
mkdir internal\timer 2>nul
mkdir web\static\css 2>nul
mkdir web\static\js 2>nul
mkdir web\templates 2>nul

echo [OK] Struktura katalogów utworzona
echo.

REM Sprawdź czy Go jest zainstalowane
echo Sprawdzanie instalacji Go...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo [BŁĄD] Go nie jest zainstalowane!
    echo Pobierz Go z: https://golang.org/dl/
    pause
    exit /b 1
)
echo [OK] Go jest zainstalowane
echo.

REM Inicjalizacja modułu Go jeśli go.mod nie istnieje
if not exist go.mod (
    echo Inicjalizacja modułu Go...
    go mod init recorder-server
    echo [OK] Moduł Go zainicjalizowany
    echo.
)

REM Sprawdź czy wszystkie pliki są na miejscu
echo Sprawdzanie plików...

set MISSING=0

if not exist "main.go" (
    echo [BRAK] main.go
    set MISSING=1
)
if not exist "config\config.go" (
    echo [BRAK] config\config.go
    set MISSING=1
)
if not exist "internal\models\models.go" (
    echo [BRAK] internal\models\models.go
    set MISSING=1
)
if not exist "internal\handlers\camera_handler.go" (
    echo [BRAK] internal\handlers\camera_handler.go
    set MISSING=1
)
if not exist "internal\handlers\obs_handler.go" (
    echo [BRAK] internal\handlers\obs_handler.go
    set MISSING=1
)
if not exist "internal\handlers\page_handler.go" (
    echo [BRAK] internal\handlers\page_handler.go
    set MISSING=1
)
if not exist "internal\handlers\timer_handler.go" (
    echo [BRAK] internal\handlers\timer_handler.go
    set MISSING=1
)
if not exist "internal\services\obs_client.go" (
    echo [BRAK] internal\services\obs_client.go
    set MISSING=1
)
if not exist "internal\services\socketio_service.go" (
    echo [BRAK] internal\services\socketio_service.go
    set MISSING=1
)
if not exist "internal\services\timer_service.go" (
    echo [BRAK] internal\services\timer_service.go
    set MISSING=1
)
if not exist "internal\state\app_state.go" (
    echo [BRAK] internal\state\app_state.go
    set MISSING=1
)
if not exist "internal\timer\models.go" (
    echo [BRAK] internal\timer\models.go
    set MISSING=1
)
if not exist "internal\timer\formatter.go" (
    echo [BRAK] internal\timer\formatter.go
    set MISSING=1
)
if not exist "internal\timer\engine.go" (
    echo [BRAK] internal\timer\engine.go
    set MISSING=1
)
if not exist "web\templates\index.html" (
    echo [BRAK] web\templates\index.html
    set MISSING=1
)
if not exist "web\static\css\style.css" (
    echo [BRAK] web\static\css\style.css
    set MISSING=1
)
if not exist "web\static\js\app.js" (
    echo [BRAK] web\static\js\app.js
    set MISSING=1
)

if %MISSING% equ 1 (
    echo.
    echo [UWAGA] Brakuje niektórych plików!
    echo Skopiuj wszystkie pliki do odpowiednich katalogów.
    echo.
    pause
    exit /b 1
)

echo [OK] Wszystkie pliki na miejscu
echo.

REM Pobierz zależności
echo Pobieranie zależności Go...
go mod tidy
if %errorlevel% neq 0 (
    echo [BŁĄD] Nie udało się pobrać zależności
    pause
    exit /b 1
)
echo [OK] Zależności pobrane
echo.

REM Pobierz Socket.IO Client
echo Pobieranie Socket.IO Client v4.5.4...
curl --version >nul 2>&1
if %errorlevel% equ 0 (
    curl -L -o web\static\js\socket.io.min.js https://cdn.socket.io/4.5.4/socket.io.min.js
    if %errorlevel% equ 0 (
        echo [OK] Socket.IO Client pobrany
    ) else (
        echo [UWAGA] Nie udało się pobrać Socket.IO automatycznie
        echo Pobierz ręcznie z: https://cdn.socket.io/4.5.4/socket.io.min.js
        echo Zapisz jako: web\static\js\socket.io.min.js
    )
) else (
    echo [UWAGA] curl nie jest dostępny
    echo Pobierz Socket.IO Client ręcznie z:
    echo https://cdn.socket.io/4.5.4/socket.io.min.js
    echo Zapisz jako: web\static\js\socket.io.min.js
)
echo.

REM Zbuduj aplikację
echo Budowanie aplikacji...
go build -o recorder-server.exe
if %errorlevel% neq 0 (
    echo [BŁĄD] Nie udało się zbudować aplikacji
    pause
    exit /b 1
)
echo [OK] Aplikacja zbudowana: recorder-server.exe
echo.

echo ========================================
echo  Setup zakończony pomyślnie!
echo ========================================
echo.
echo Aby uruchomić aplikację:
echo   recorder-server.exe
echo.
echo Panel WWW dostępny na:
echo   http://localhost:8080
echo.
echo Pamiętaj o:
echo - Włączeniu OBS Studio
echo - Aktywacji WebSocket Server w OBS
echo - Ustawieniu portu 4445 w OBS
echo.
pause