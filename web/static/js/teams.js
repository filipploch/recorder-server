// teams.js - Zarządzanie zespołami i strojami

let currentTeamId = null;
let currentTempId = null;
let isEditingTemp = false;
let availableLogos = [];
let hasTeamsUrl = false;

// Kits data structure
let kitsData = {
    1: ['#ffffff'],        // Komplet 1 - domowy (biały)
    2: ['#000000'],        // Komplet 2 - wyjazdowy (czarny)
    3: ['#66ff73']         // Komplet 3 - rezerwowy (zielony)
};

// Color picker state
let currentEditingKit = null;
let currentEditingSegment = null;

// ============ INITIALIZATION ============

document.addEventListener('DOMContentLoaded', function() {
    loadLogos();
    checkCompetitionScraper();
    loadAll();
    initializeKits();
});

// ============ KIT EDITOR FUNCTIONS ============

function initializeKits() {
    // Initialize with default colors
    kitsData = {
        1: ['#ffffff'],
        2: ['#000000'],
        3: ['#66ff73']
    };
    renderAllKits();
}

function renderAllKits() {
    renderKit(1);
    renderKit(2);
    renderKit(3);
}

function renderKit(kitType) {
    const container = document.getElementById(`kit${kitType}-segments`);
    const colors = kitsData[kitType];
    
    container.innerHTML = '';
    
    colors.forEach((color, index) => {
        const segment = document.createElement('div');
        segment.className = 'kit-segment';
        segment.style.backgroundColor = color;
        segment.onclick = () => openColorPicker(kitType, index);
        container.appendChild(segment);
    });
    
    // Update button states
    updateKitButtons(kitType);
}

function updateKitButtons(kitType) {
    const colors = kitsData[kitType];
    const removeBtn = document.getElementById(`kit${kitType}-remove`);
    const addBtn = document.getElementById(`kit${kitType}-add`);
    
    removeBtn.disabled = colors.length <= 1;
    addBtn.disabled = colors.length >= 5;
}

function addKitColor(kitType) {
    if (kitsData[kitType].length >= 5) return;
    
    // Add new color (copy last color)
    const lastColor = kitsData[kitType][kitsData[kitType].length - 1];
    kitsData[kitType].push(lastColor);
    
    renderKit(kitType);
}

function removeKitColor(kitType) {
    if (kitsData[kitType].length <= 1) return;
    
    // Remove last color
    kitsData[kitType].pop();
    
    renderKit(kitType);
}

function openColorPicker(kitType, segmentIndex) {
    currentEditingKit = kitType;
    currentEditingSegment = segmentIndex;
    
    const currentColor = kitsData[kitType][segmentIndex];
    document.getElementById('colorPicker').value = currentColor;
    document.getElementById('colorPickerModal').classList.add('active');
}

function confirmColorPicker() {
    const newColor = document.getElementById('colorPicker').value;
    kitsData[currentEditingKit][currentEditingSegment] = newColor;
    
    renderKit(currentEditingKit);
    closeColorPicker();
}

function cancelColorPicker() {
    closeColorPicker();
}

function closeColorPicker() {
    document.getElementById('colorPickerModal').classList.remove('active');
    currentEditingKit = null;
    currentEditingSegment = null;
}

function loadKitsFromTeam(team) {
    // Load kits from team data
    if (team.kits && team.kits.length === 3) {
        team.kits.forEach(kit => {
            if (kit.kit_colors && kit.kit_colors.length > 0) {
                // Sort by ColorOrder
                const sortedColors = kit.kit_colors.sort((a, b) => a.color_order - b.color_order);
                kitsData[kit.type] = sortedColors.map(kc => kc.color);
            }
        });
    }
    renderAllKits();
}

function getKitsForSave() {
    // Convert kitsData to API format as a map
    return {
        "1": kitsData[1],
        "2": kitsData[2],
        "3": kitsData[3]
    };
}

// ============ API FUNCTIONS ============

function checkCompetitionScraper() {
    fetch('/api/competition/current')
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success' && data.competition) {
                try {
                    const variable = JSON.parse(data.competition.variable);
                    if (variable.scraper && variable.scraper.teams_url) {
                        hasTeamsUrl = true;
                        document.getElementById('scrapeButton').style.display = 'inline-block';
                    }
                } catch (e) {
                    console.error('Error parsing Variable:', e);
                }
            }
        })
        .catch(error => console.error('Error:', error));
}

function loadLogos() {
    fetch('/api/logos')
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success') {
                availableLogos = data.logos;
                updateLogoSelect();
            }
        })
        .catch(error => console.error('Error loading logos:', error));
}

function updateLogoSelect() {
    const select = document.getElementById('logo');
    select.innerHTML = '<option value="">-- Wybierz logo --</option>';
    
    availableLogos.forEach(logo => {
        const option = document.createElement('option');
        option.value = logo.url;
        option.textContent = logo.name;
        select.appendChild(option);
    });
}

function updateLogoPreview() {
    const select = document.getElementById('logo');
    const preview = document.getElementById('logoPreview');
    
    if (select.value) {
        preview.src = select.value;
        preview.style.display = 'block';
    } else {
        preview.style.display = 'none';
    }
}

function loadAll() {
    loadTeams();
    loadTempTeams();
}

function loadTeams() {
    document.getElementById('loading').style.display = 'block';
    document.getElementById('teamsTable').style.display = 'none';
    document.getElementById('emptyState').style.display = 'none';

    fetch('/api/teams')
        .then(response => response.json())
        .then(data => {
            document.getElementById('loading').style.display = 'none';
            
            if (data.status === 'success' && data.teams && data.teams.length > 0) {
                renderTeams(data.teams);
            } else {
                document.getElementById('emptyState').style.display = 'block';
            }
        })
        .catch(error => {
            document.getElementById('loading').style.display = 'none';
            console.error('Error:', error);
            showAlert('Błąd ładowania zespołów', 'error');
        });
}

function loadTempTeams() {
    document.getElementById('tempLoading').style.display = 'block';
    document.getElementById('tempTeamsTable').style.display = 'none';
    document.getElementById('tempEmptyState').style.display = 'none';

    fetch('/api/teams/temp')
        .then(response => response.json())
        .then(data => {
            document.getElementById('tempLoading').style.display = 'none';
            
            if (data.status === 'success') {
                if (data.teams && data.teams.length > 0) {
                    document.getElementById('tempTeamsSection').style.display = 'block';
                    renderTempTeams(data.teams);
                    renderTempStats(data.statistics);
                    
                    if (data.statistics && data.statistics.complete > 0) {
                        document.getElementById('importAllBtn').style.display = 'inline-block';
                    } else {
                        document.getElementById('importAllBtn').style.display = 'none';
                    }
                } else {
                    document.getElementById('tempTeamsSection').style.display = 'none';
                    document.getElementById('importAllBtn').style.display = 'none';
                }
            } else {
                document.getElementById('tempTeamsSection').style.display = 'none';
                document.getElementById('importAllBtn').style.display = 'none';
            }
        })
        .catch(error => {
            document.getElementById('tempLoading').style.display = 'none';
            console.error('Error loading temp teams:', error);
            showAlert('Błąd ładowania tymczasowych zespołów', 'error');
        });
}

function renderTempStats(stats) {
    const container = document.getElementById('tempStats');
    container.innerHTML = `
        <div class="stat-card">
            <h3>Łącznie</h3>
            <div class="value">${stats.total || 0}</div>
        </div>
        <div class="stat-card">
            <h3>Kompletne</h3>
            <div class="value" style="color: #28a745;">${stats.complete || 0}</div>
        </div>
        <div class="stat-card">
            <h3>Niekompletne</h3>
            <div class="value" style="color: #ffc107;">${stats.incomplete || 0}</div>
        </div>
    `;
}

function renderTempTeams(teams) {
    const tbody = document.getElementById('tempTeamsBody');
    tbody.innerHTML = '';

    teams.forEach(team => {
        const tr = document.createElement('tr');
        
        const isComplete = team.short_name && team.name_16 && team.logo;
        const statusBadge = isComplete 
            ? '<span class="badge badge-complete">Kompletny</span>'
            : '<span class="badge badge-incomplete">Niekompletny</span>';
        
        const logoCell = team.logo 
            ? `<img src="${team.logo}" alt="${team.name}" class="team-logo">`
            : '-';
        
        const shortName = team.short_name || '-';
        const name16 = team.name_16 || '-';
        
        tr.innerHTML = `
            <td>${statusBadge}</td>
            <td class="team-name">${team.name}</td>
            <td>${shortName}</td>
            <td>${name16}</td>
            <td>${logoCell}</td>
            <td><small>${team.source}</small></td>
            <td class="action-buttons">
                <button class="btn btn-primary btn-sm" onclick="openEditTempModal('${team.temp_id}')">
                    ✏️ Edytuj
                </button>
                ${isComplete ? `
                    <button class="btn btn-success btn-sm" onclick="importTempTeam('${team.temp_id}')">
                        💾 Importuj
                    </button>
                ` : ''}
                <button class="btn btn-danger btn-sm" onclick="deleteTempTeam('${team.temp_id}', '${team.name}')">
                    🗑️ Usuń
                </button>
            </td>
        `;
        
        tbody.appendChild(tr);
    });

    document.getElementById('tempTeamsTable').style.display = 'table';
}

function renderTeams(teams) {
    const tbody = document.getElementById('teamsBody');
    tbody.innerHTML = '';

    teams.forEach(team => {
        const tr = document.createElement('tr');
        
        const logoUrl = team.logo || '/static/images/default-logo.png';
        const linkCell = team.link ? `<a href="${team.link}" target="_blank">🔗</a>` : '-';
        
        // Render kit previews
        let kitsHtml = '<div class="kit-preview">';
        if (team.kits && team.kits.length === 3) {
            team.kits.sort((a, b) => a.type - b.type).forEach(kit => {
                if (kit.kit_colors && kit.kit_colors.length > 0) {
                    const sortedColors = kit.kit_colors.sort((a, b) => a.color_order - b.color_order);
                    sortedColors.forEach(kc => {
                        kitsHtml += `<div class="kit-preview-segment" style="background-color: ${kc.color}"></div>`;
                    });
                    kitsHtml += '<span style="margin: 0 5px;">|</span>';
                }
            });
        }
        kitsHtml += '</div>';
        
        tr.innerHTML = `
            <td><img src="${logoUrl}" alt="${team.name}" class="team-logo"></td>
            <td class="team-name">${team.name}</td>
            <td><span class="team-short">${team.short_name}</span></td>
            <td>${team.name_16}</td>
            <td>${kitsHtml}</td>
            <td>${linkCell}</td>
            <td class="action-buttons">
                <button class="btn btn-primary btn-sm" onclick="openEditModal(${team.id})">
                    ✏️ Edytuj
                </button>
                <button class="btn btn-danger btn-sm" onclick="deleteTeam(${team.id}, '${team.name}')">
                    🗑️ Usuń
                </button>
            </td>
        `;
        
        tbody.appendChild(tr);
    });

    document.getElementById('teamsTable').style.display = 'table';
}

// ============ MODAL FUNCTIONS ============

function openCreateModal() {
    currentTeamId = null;
    currentTempId = null;
    isEditingTemp = false;
    document.getElementById('modalTitle').textContent = 'Dodaj nowy zespół';
    document.getElementById('teamId').value = '';
    document.getElementById('teamForm').reset();
    clearValidationErrors();
    document.getElementById('logoPreview').style.display = 'none';
    
    // Reset kits to default
    initializeKits();
    
    document.getElementById('teamModal').classList.add('active');
}

function openEditTempModal(tempId) {
    currentTempId = tempId;
    currentTeamId = null;
    isEditingTemp = true;
    document.getElementById('modalTitle').textContent = 'Edytuj zespół tymczasowy';
    clearValidationErrors();
    
    // Reset kits to default
    initializeKits();

    fetch(`/api/teams/temp/${tempId}`)
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success') {
                const team = data.team;
                document.getElementById('teamId').value = '';
                document.getElementById('name').value = team.name || '';
                document.getElementById('short_name').value = team.short_name || '';
                document.getElementById('name_16').value = team.name_16 || '';
                document.getElementById('logo').value = team.logo || '';
                document.getElementById('link').value = team.link || '';
                document.getElementById('foreign_id').value = team.foreign_id || '';
                
                updateLogoPreview();
                document.getElementById('teamModal').classList.add('active');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showAlert('Błąd ładowania danych zespołu', 'error');
        });
}

function openEditModal(teamId) {
    currentTeamId = teamId;
    currentTempId = null;
    isEditingTemp = false;
    document.getElementById('modalTitle').textContent = 'Edytuj zespół';
    clearValidationErrors();

    fetch(`/api/teams/${teamId}`)
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success') {
                const team = data.team;
                document.getElementById('teamId').value = team.id;
                document.getElementById('name').value = team.name || '';
                document.getElementById('short_name').value = team.short_name || '';
                document.getElementById('name_16').value = team.name_16 || '';
                document.getElementById('logo').value = team.logo || '';
                document.getElementById('link').value = team.link || '';
                document.getElementById('foreign_id').value = team.foreign_id || '';
                
                // Load kits
                loadKitsFromTeam(team);
                
                updateLogoPreview();
                document.getElementById('teamModal').classList.add('active');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showAlert('Błąd ładowania danych zespołu', 'error');
        });
}

function closeModal() {
    document.getElementById('teamModal').classList.remove('active');
    document.getElementById('teamForm').reset();
    clearValidationErrors();
    currentTeamId = null;
    currentTempId = null;
    isEditingTemp = false;
    document.getElementById('logoPreview').style.display = 'none';
    initializeKits();
}

function saveTeam() {
    clearValidationErrors();

    const teamData = {
        name: document.getElementById('name').value,
        short_name: document.getElementById('short_name').value,
        name_16: document.getElementById('name_16').value,
        logo: document.getElementById('logo').value,
        link: document.getElementById('link').value || null,
        foreign_id: document.getElementById('foreign_id').value || null,
        kits: getKitsForSave() // Include kits data
    };

    console.log("teamData", teamData);

    let url, method;
    
    if (isEditingTemp) {
        url = `/api/teams/temp/${currentTempId}`;
        method = 'PUT';
        // Dla temp teams zawsze wysyłamy kits
    } else if (currentTeamId) {
        url = `/api/teams/${currentTeamId}`;
        method = 'PUT';
    } else {
        url = '/api/teams';
        method = 'POST';
    }
    console.log("501");
    console.log("url", url);
    console.log("method", method);
    console.log("currentTempId", currentTempId);
    console.log("currentTeamId", currentTeamId);
    fetch(url, {
        method: method,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(teamData)
    })
    .then(response => {
        // Sprawdź status odpowiedzi
        if (!response.ok) {
            return response.text().then(text => {
                throw new Error(`HTTP ${response.status}: ${text}`);
            });
        }
        return response.json();
    })
    .then(data => {
        console.log("data 519", data);
        if (data.status === 'success') {
            closeModal();
            
            if (data.imported) {
                // Drużyna została zaimportowana do bazy
                showAlert('✓ Drużyna została zaimportowana do bazy danych', 'success');
                loadAll(); // Odśwież obie tabele
            } else if (isEditingTemp) {
                // Drużyna zapisana w pliku tymczasowym
                showAlert('✓ Zmiany zapisane w pliku tymczasowym', 'success');
                loadTempTeams();
            } else {
                // Normalna edycja/tworzenie
                showAlert(data.message || 'Zespół zapisany pomyślnie', 'success');
                loadTeams();
            }
        } else if (data.status === 'validation_error') {
            displayValidationErrors(data.errors);
        } else {
            showAlert(data.error || 'Błąd zapisywania zespołu', 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showAlert('Błąd: ' + error.message, 'error');
    });
}

// ============ OTHER FUNCTIONS ============

function scrapeTeams() {
    const btn = document.getElementById('scrapeButton');
    btn.disabled = true;
    btn.textContent = '⏳ Pobieranie...';

    fetch('/api/scrape/teams', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            competition_id: 1
        })
    })
    .then(response => response.json())
    .then(data => {
        btn.disabled = false;
        btn.textContent = '🌐 Pobierz zespoły';
        
        if (data.status === 'success') {
            showAlert(`✓ ${data.message}. Nowych: ${data.new}, Łącznie w pliku: ${data.total}`, 'success');
            loadTempTeams();
        } else {
            showAlert('✗ ' + (data.error || 'Błąd pobierania zespołów'), 'error');
        }
    })
    .catch(error => {
        btn.disabled = false;
        btn.textContent = '🌐 Pobierz zespoły';
        console.error('Error:', error);
        showAlert('✗ Błąd połączenia', 'error');
    });
}

function importTempTeam(tempId) {
    if (!confirm('Czy na pewno chcesz zaimportować ten zespół do bazy danych?')) {
        return;
    }

    fetch(`/api/teams/import/${tempId}`, {
        method: 'POST'
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            showAlert('✓ ' + (data.message || 'Zespół zaimportowany'), 'success');
            loadAll();
        } else {
            showAlert('✗ ' + (data.error || 'Błąd importowania'), 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showAlert('✗ Błąd połączenia', 'error');
    });
}

function importAllComplete() {
    if (!confirm('Czy na pewno chcesz zaimportować wszystkie kompletne zespoły?')) {
        return;
    }

    const btn = document.getElementById('importAllBtn');
    btn.disabled = true;
    btn.textContent = '⏳ Importowanie...';

    fetch('/api/teams/import-all', {
        method: 'POST'
    })
    .then(response => response.json())
    .then(data => {
        btn.disabled = false;
        btn.textContent = '💾 Importuj wszystkie kompletne';
        
        if (data.status === 'success') {
            let message = `✓ Zaimportowano ${data.imported} zespołów`;
            if (data.errors && data.errors.length > 0) {
                message += `\n⚠️ Błędy: ${data.errors.length}`;
            }
            showAlert(message, 'success');
            loadAll();
        } else {
            showAlert('✗ Błąd importowania', 'error');
        }
    })
    .catch(error => {
        btn.disabled = false;
        btn.textContent = '💾 Importuj wszystkie kompletne';
        console.error('Error:', error);
        showAlert('✗ Błąd połączenia', 'error');
    });
}

function deleteTempTeam(tempId, teamName) {
    if (!confirm(`Czy na pewno chcesz usunąć tymczasowy zespół "${teamName}"?`)) {
        return;
    }

    fetch(`/api/teams/temp/${tempId}`, {
        method: 'DELETE'
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            loadTempTeams();
            showAlert(data.message || 'Zespół usunięty', 'success');
        } else {
            showAlert(data.error || 'Błąd usuwania', 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showAlert('Błąd połączenia', 'error');
    });
}

function deleteTeam(teamId, teamName) {
    if (!confirm(`Czy na pewno chcesz usunąć zespół "${teamName}"?`)) {
        return;
    }

    fetch(`/api/teams/${teamId}`, {
        method: 'DELETE'
    })
    .then(response => response.json())
    .then(data => {
        if (data.status === 'success') {
            loadTeams();
            showAlert(data.message || 'Zespół usunięty', 'success');
        } else {
            showAlert(data.error || 'Błąd usuwania zespołu', 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showAlert('Błąd połączenia z serwerem', 'error');
    });
}

function displayValidationErrors(errors) {
    errors.forEach(error => {
        const groupElement = document.getElementById(`group-${error.field}`);
        const errorElement = document.getElementById(`error-${error.field}`);
        
        if (groupElement && errorElement) {
            groupElement.classList.add('error');
            errorElement.textContent = error.message;
        }
    });
}

function clearValidationErrors() {
    const errorGroups = document.querySelectorAll('.form-group.error');
    errorGroups.forEach(group => {
        group.classList.remove('error');
    });
    
    const errorMessages = document.querySelectorAll('.error-message');
    errorMessages.forEach(msg => {
        msg.textContent = '';
    });
}

function showAlert(message, type) {
    const alert = document.getElementById('alert');
    alert.textContent = message;
    alert.className = 'alert active';
    
    if (type === 'success') {
        alert.classList.add('alert-success');
    } else if (type === 'info') {
        alert.classList.add('alert-info');
    } else {
        alert.classList.add('alert-error');
    }

    setTimeout(() => {
        alert.classList.remove('active');
    }, 5000);
}

// Close modal on outside click
window.onclick = function(event) {
    const modal = document.getElementById('teamModal');
    const colorModal = document.getElementById('colorPickerModal');
    
    if (event.target === modal) {
        closeModal();
    }
    if (event.target === colorModal) {
        closeColorPicker();
    }
}

// Close modal on ESC key
document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape') {
        closeModal();
        closeColorPicker();
    }
});
