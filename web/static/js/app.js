// Połączenie Socket.IO
const socket = io();

// Event: połączono z serwerem
socket.on('connect', function() {
    console.log('Połączono z serwerem Socket.IO');
    console.log('Socket ID:', socket.id);
    console.log('Socket connected:', socket.connected);
    updateStatus('Połączono z serwerem ✓');
});

// Event: rozłączono z serwerem
socket.on('disconnect', function() {
    console.log('Rozłączono z serwerem');
    updateStatus('Rozłączono z serwerem ✗');
});

// Funkcje pomocnicze

function updateStatus(message) {
    document.getElementById('status').innerHTML = message;
}

function updateOBSStatus(connected, message) {
    const indicator = document.getElementById('obs-indicator');
    const statusText = document.getElementById('obs-status');
    
    if (connected) {
        indicator.className = 'obs-status obs-connected';
        statusText.textContent = 'OBS: ' + (message || 'Połączono ✓');
    } else {
        indicator.className = 'obs-status obs-disconnected';
        statusText.textContent = 'OBS: ' + (message || 'Rozłączono ✗');
    }
}

function getCheckedCameras() {
    const cameras = ['camera_main', 'camera_center', 'camera_left', 'camera_right'];
    const active = [];
    const inactive = [];
    
    cameras.forEach(camera => {
        const checkbox = document.getElementById(camera);
        if (checkbox && checkbox.checked) {
            active.push(camera);
        } else {
            inactive.push(camera);
        }
    });
    
    return { active, inactive };
}

// API: Kamery

function startRecording() {
    const cameras = getCheckedCameras();
    
    fetch('/api/start-recording', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            active_cameras: cameras.active,
            inactive_cameras: cameras.inactive
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            updateStatus('✓ Rozpoczęto nagrywanie kamer');
        } else {
            updateStatus('✗ Błąd: ' + (data.error || 'Nieznany błąd'));
        }
    })
    .catch(error => {
        console.error('Błąd:', error);
        updateStatus('✗ Błąd połączenia');
    });
}

function stopRecording() {
    fetch('/api/stop-recording', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ cameras: [] })
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            updateStatus('✓ Zatrzymano nagrywanie kamer');
        } else {
            updateStatus('✗ Błąd: ' + (data.error || 'Nieznany błąd'));
        }
    })
    .catch(error => {
        console.error('Błąd:', error);
        updateStatus('✗ Błąd połączenia');
    });
}

function getStatus() {
    fetch('/api/status')
        .then(response => response.json())
        .then(data => {
            const statusText = data.record_status ? 
                '<strong>NAGRYWANIE AKTYWNE</strong>' : 
                '<strong>ZATRZYMANE</strong>';
            
            updateStatus(
                statusText +
                '<br>Aktywne kamery: ' + (data.active_cameras.length > 0 ? 
                    data.active_cameras.join(', ') : 'brak') +
                '<br>Nieaktywne kamery: ' + (data.inactive_cameras.length > 0 ? 
                    data.inactive_cameras.join(', ') : 'brak')
            );
        })
        .catch(error => {
            console.error('Błąd:', error);
            updateStatus('✗ Błąd pobierania statusu');
        });
}

// API: OBS Studio

function obsStartRecording() {
    fetch('/api/obs/start-recording', { method: 'POST' })
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success') {
                updateOBSStatus(true, 'Rozpoczęto nagrywanie ✓');
            } else {
                alert('Błąd OBS: ' + (data.error || 'Nieznany błąd'));
            }
        })
        .catch(error => {
            console.error('Błąd:', error);
            alert('Błąd połączenia z OBS');
        });
}

function obsStopRecording() {
    fetch('/api/obs/stop-recording', { method: 'POST' })
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success') {
                updateOBSStatus(true, 'Zatrzymano nagrywanie ✓');
            } else {
                alert('Błąd OBS: ' + (data.error || 'Nieznany błąd'));
            }
        })
        .catch(error => {
            console.error('Błąd:', error);
            alert('Błąd połączenia z OBS');
        });
}

function obsGetStatus() {
    fetch('/api/obs/status')
        .then(response => response.json())
        .then(data => {
            if (data.connected) {
                const msg = data.recording ? 
                    'Nagrywanie aktywne ✓' : 
                    'Gotowy (nie nagrywa)';
                updateOBSStatus(true, msg);
            } else {
                updateOBSStatus(false, 'Rozłączono ✗');
            }
        })
        .catch(error => {
            console.error('Błąd:', error);
            updateOBSStatus(false, 'Błąd połączenia ✗');
        });
}

function obsGetScenes() {
    fetch('/api/obs/scenes')
        .then(response => response.json())
        .then(data => {
            const container = document.getElementById('scenes-container');
            
            if (data.status === 'success' && data.scenes) {
                container.innerHTML = '<h4>Dostępne sceny OBS:</h4><div class="button-group" id="scene-buttons"></div>';
                const buttonContainer = document.getElementById('scene-buttons');
                
                data.scenes.forEach(scene => {
                    const btn = document.createElement('button');
                    btn.className = 'btn btn-scene';
                    btn.textContent = scene;
                    btn.onclick = () => obsSetScene(scene);
                    buttonContainer.appendChild(btn);
                });
            } else {
                container.innerHTML = '<p style="color: red;">Błąd: ' + 
                    (data.error || 'Nie można pobrać scen') + '</p>';
            }
        })
        .catch(error => {
            console.error('Błąd:', error);
            document.getElementById('scenes-container').innerHTML = 
                '<p style="color: red;">Błąd połączenia</p>';
        });
}

function obsSetScene(sceneName) {
    fetch('/api/obs/set-scene', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ scene_name: sceneName })
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            updateOBSStatus(true, 'Zmieniono scenę: ' + sceneName + ' ✓');
        } else {
            alert('Błąd zmiany sceny: ' + (data.error || 'Nieznany błąd'));
        }
    })
    .catch(error => {
        console.error('Błąd:', error);
        alert('Błąd połączenia');
    });
}

// Auto-refresh statusu co 5 sekund
setInterval(getStatus, 5000);
setInterval(obsGetStatus, 5000);

// Auto-refresh stopera co 0.5 sekundy (backup na wypadek problemów z Socket.IO)
setInterval(timerGetState, 500);

// Pobierz status przy załadowaniu strony
window.addEventListener('DOMContentLoaded', function() {
    getStatus();
    obsGetStatus();
    timerGetState();
});

// Socket.IO event: timer update
socket.on('timer_update', function(data) {
    console.log('Timer update received:', data);
    updateTimerDisplay(data);
});

// Timer API

function timerStart() {
    const direction = document.getElementById('timer-direction').value;
    const broadcastPrecision = document.getElementById('timer-broadcast-precision').value;
    const maxDurationInput = document.getElementById('timer-max-duration').value;
    const stopBehavior = document.getElementById('timer-stop-behavior').value;
    
    const payload = {
        measurement_precision: 'ms', // zawsze pomiar w ms dla dokładności
        broadcast_precision: broadcastPrecision,
        direction: direction,
        stop_behavior: stopBehavior
    };
    
    // Dodaj max_duration jeśli podano
    if (maxDurationInput && maxDurationInput.trim() !== '') {
        const maxDuration = parseInt(maxDurationInput);
        if (!isNaN(maxDuration) && maxDuration > 0) {
            payload.max_duration = maxDuration;
        }
    }
    
    console.log('=== TIMER START ===');
    console.log('Starting timer with payload:', payload);
    
    fetch('/api/timer/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
    })
    .then(response => {
        console.log('Timer start response status:', response.status);
        return response.json();
    })
    .then(data => {
        console.log('Timer start response:', data);
        if (data.status === 'success') {
            console.log('✓ Timer: Uruchomiono/wznowiono');
        } else {
            console.error('✗ Błąd stopera:', data.error);
            alert('Błąd stopera: ' + (data.error || 'Nieznany błąd'));
        }
    })
    .catch(error => {
        console.error('✗ Błąd połączenia:', error);
        alert('Błąd połączenia ze stoperem');
    });
}

function timerPause() {
    console.log('=== TIMER PAUSE ===');
    fetch('/api/timer/pause', { method: 'POST' })
        .then(response => {
            console.log('Timer pause response status:', response.status);
            return response.json();
        })
        .then(data => {
            console.log('Timer pause response:', data);
            if (data.status === 'success') {
                console.log('✓ Timer: Zapauzowano');
            }
        })
        .catch(error => console.error('✗ Błąd:', error));
}

function timerReset() {
    fetch('/api/timer/reset', { method: 'POST' })
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success') {
                console.log('Timer: Zresetowano');
                // Odśwież wyświetlacz
                timerGetState();
            }
        })
        .catch(error => console.error('Błąd:', error));
}

function timerGetState() {
    fetch('/api/timer/state')
        .then(response => response.json())
        .then(data => {
            console.log('Timer state:', data);
            // Użyj tej samej funkcji co dla Socket.IO updates
            updateTimerDisplay(data);
        })
        .catch(error => console.error('Błąd:', error));
}

function timerQuickStart(maxDuration, direction) {
    // Ustaw wartości w formularzu
    document.getElementById('timer-direction').value = direction;
    document.getElementById('timer-broadcast-precision').value = 'ds';
    
    if (maxDuration) {
        document.getElementById('timer-max-duration').value = maxDuration;
        document.getElementById('timer-stop-behavior').value = 'auto';
    } else {
        document.getElementById('timer-max-duration').value = '';
    }
    
    // Uruchom
    timerStart();
}

function updateTimerDisplay(data) {
    console.log('Updating timer display with:', data);
    const display = document.getElementById('timer-display');
    const info = document.getElementById('timer-info');
    
    if (!display || !info) {
        console.error('Timer display elements not found!');
        return;
    }
    
    // Aktualizuj wyświetlacz
    const timeToDisplay = data.formatted_time || '00:00';
    display.textContent = timeToDisplay;
    console.log('Set display text to:', timeToDisplay);
    
    // Dodaj klasę overflow jeśli przekroczono czas
    if (data.is_overflow) {
        display.classList.add('overflow');
    } else {
        display.classList.remove('overflow');
    }
    
    // Aktualizuj informacje
    let infoText = '';
    
    // Sprawdź różne warianty pola running
    const isRunning = data.running === true || data.Running === true;
    
    if (isRunning) {
        infoText = '⏱️ Stoper działa';
    } else {
        infoText = '⏸️ Stoper zatrzymany';
    }
    
    const maxDuration = data.max_duration_ms || data.MaxDurationMs;
    if (maxDuration) {
        const maxSec = Math.floor(maxDuration / 1000);
        infoText += ` | Max: ${maxSec}s`;
    }
    
    const isOverflow = data.is_overflow || data.IsOverflow;
    if (isOverflow) {
        infoText += ' | ⚠️ PRZEKROCZONO CZAS';
    }
    
    info.textContent = infoText;
    console.log('Updated timer info:', infoText);
}