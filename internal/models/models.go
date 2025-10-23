package models

import (
	"time"
	"gorm.io/gorm"
)

// Team - drużyna
type Team struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Link      string `json:"link"`
	ForeignID *string `json:"foreign_id"` // nullable
	Name      string `gorm:"not null" json:"name"`
	ShortName string `gorm:"size:3;not null;unique" json:"short_name"`
	Name16    string `gorm:"size:16;not null;unique" json:"name_16"`
	Logo      string `json:"logo"`
	
	// Relacje
	Players []Player `gorm:"foreignKey:TeamID" json:"players,omitempty"`
	Coaches []Coach  `gorm:"foreignKey:TeamID" json:"coaches,omitempty"`
	Kits    []Kit    `gorm:"foreignKey:TeamID" json:"kits,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// PlayerRole - pozycja zawodnika
type PlayerRole struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"not null" json:"name"`
	ShortName string `json:"short_name"`
	
	// Relacje
	Players []Player `gorm:"foreignKey:PlayerRoleID" json:"players,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Player - zawodnik
type Player struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	Link         string  `json:"link"`
	ForeignID    *string `json:"foreign_id"` // nullable
	FirstName    string  `gorm:"not null" json:"first_name"`
	LastName     string  `gorm:"not null" json:"last_name"`
	Number       string  `json:"number"` // String bo może być "18+5"
	PlayerRoleID uint    `json:"player_role_id"`
	TeamID       uint    `json:"team_id"`
	IsCaptain    bool    `gorm:"default:false" json:"is_captain"`
	
	// Relacje
	PlayerRole PlayerRole `gorm:"foreignKey:PlayerRoleID" json:"player_role,omitempty"`
	Team       Team       `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// CoachRole - rola trenera
type CoachRole struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"not null" json:"name"`
	ShortName string `json:"short_name"`
	
	// Relacje
	Coaches []Coach `gorm:"foreignKey:CoachRoleID" json:"coaches,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Coach - trener
type Coach struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	FirstName   string `gorm:"not null" json:"first_name"`
	LastName    string `gorm:"not null" json:"last_name"`
	CoachRoleID uint   `json:"coach_role_id"`
	TeamID      uint   `json:"team_id"`
	
	// Relacje
	CoachRole CoachRole `gorm:"foreignKey:CoachRoleID" json:"coach_role,omitempty"`
	Team      Team      `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// RefereeRole - rola sędziego
type RefereeRole struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"not null" json:"name"`
	ShortName string `json:"short_name"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Referee - sędzia
type Referee struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	FirstName string `gorm:"not null" json:"first_name"`
	LastName  string `gorm:"not null" json:"last_name"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Field - boisko/hala
type Field struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Name   string `gorm:"not null" json:"name"`
	City   string `json:"city"`
	Street string `json:"street"`
	
	// Relacje
	Games []Game `gorm:"foreignKey:FieldID" json:"games,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Competition - rozgrywki (liga, puchar, mistrzostwa)
type Competition struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	Name            string `gorm:"not null" json:"name"`
	Season          string `json:"season"` // np. "2024/2025"
	CompetitionType string `json:"competition_type"` // "league", "cup", "tournament"
	
	// Relacje
	Stages []Stage `gorm:"foreignKey:CompetitionID" json:"stages,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Stage - etap rozgrywek (np. faza grupowa, 1/8 finału)
type Stage struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	CompetitionID uint   `json:"competition_id"`
	Name          string `gorm:"not null" json:"name"` // np. "Grupa A", "1/8 finału", "Kolejka 5"
	StageOrder    int    `json:"stage_order"` // kolejność etapu
	ParentStageID *uint  `json:"parent_stage_id"` // nullable - etap z którego awansowano
	
	// Relacje
	Competition Competition `gorm:"foreignKey:CompetitionID" json:"competition,omitempty"`
	ParentStage *Stage      `gorm:"foreignKey:ParentStageID" json:"parent_stage,omitempty"`
	Groups      []Group     `gorm:"foreignKey:StageID" json:"groups,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Group - grupa drużyn (liga, grupa pucharowa, para w pucharze)
type Group struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	StageID  uint   `json:"stage_id"`
	Name     string `json:"name"` // np. "Grupa A", "Para 1", "Tabela ligowa"
	IsTwoLeg bool   `gorm:"default:false" json:"is_two_leg"` // czy dwumecz
	
	// Relacje
	Stage      Stage       `gorm:"foreignKey:StageID" json:"stage,omitempty"`
	GroupTeams []GroupTeam `gorm:"foreignKey:GroupID" json:"group_teams,omitempty"`
	Games      []Game      `gorm:"foreignKey:GroupID" json:"games,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GroupTeam - drużyna w grupie
type GroupTeam struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	GroupID uint `json:"group_id"`
	TeamID  uint `json:"team_id"`
	
	// Relacje
	Group Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Team  Team  `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Game - mecz
type Game struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	GroupID   *uint  `json:"group_id"` // nullable - grupa/para do której należy mecz
	FieldID   uint   `json:"field_id"`
	DateTime  string `json:"date_time"` // Format: "2025-10-17_20:45"
	LegNumber int    `gorm:"default:1" json:"leg_number"` // 1 lub 2 (dla dwumeczów)
	
	// Relacje
	Group *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Field Field  `gorm:"foreignKey:FieldID" json:"field,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GamePart - część meczu (kwarta, połowa, etc.)
type GamePart struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	Name       string  `gorm:"not null" json:"name"`
	MinPoints  *int    `json:"min_points"` // nullable
	Length     *int    `json:"length"`     // nullable - czas w sekundach
	MatchOrder int     `gorm:"unique;not null" json:"match_order"`
	TimeGroup  *uint   `json:"time_group"`  // nullable - grupa do sumowania czasów
	ActualTime int     `gorm:"default:0" json:"actual_time"` // ostatnio zapisany czas w sekundach
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TVStaffRole - rola członka ekipy TV
type TVStaffRole struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"not null" json:"name"`
	ShortName string `json:"short_name"`
	
	// Relacje
	TVStaff []TVStaff `gorm:"foreignKey:TVStaffRoleID" json:"tv_staff,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TVStaff - członek ekipy telewizyjnej
type TVStaff struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	FirstName     string `gorm:"not null" json:"first_name"`
	LastName      string `gorm:"not null" json:"last_name"`
	TVStaffRoleID uint   `json:"tv_staff_role_id"`
	
	// Relacje
	TVStaffRole TVStaffRole `gorm:"foreignKey:TVStaffRoleID" json:"tv_staff_role,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ValueType - typ wartości (bramki, faule, etc.)
type ValueType struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"not null" json:"name"`      // np: "Bramki"
	ShortName string `gorm:"not null" json:"short_name"` // np: "B"
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Camera - kamera
type Camera struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"not null" json:"name"`
	Location string `json:"location"`
	
	// Relacje
	Events  []Event  `gorm:"foreignKey:CameraID" json:"events,omitempty"`
	Replays []Replay `gorm:"foreignKey:CameraID" json:"replays,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// EventType - typ wydarzenia
type EventType struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	Name         string  `gorm:"not null" json:"name"`
	Icon         *string `json:"icon"` // nullable - URL do ikonki
	ShortName    string  `json:"short_name"`
	IsInProtocol bool    `gorm:"default:false" json:"is_in_protocol"`
	
	// Relacje
	Events []Event `gorm:"foreignKey:EventTypeID" json:"events,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Event - wydarzenie w meczu
type Event struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	EventTypeID uint   `json:"event_type_id"`
	GameID      uint   `json:"game_id"`
	GamePartID  uint   `json:"game_part_id"`
	EventTime   int    `json:"event_time"` // czas w sekundach
	CameraID    uint   `json:"camera_id"`
	TeamID      *uint  `json:"team_id"`   // nullable
	PlayerID    *uint  `json:"player_id"` // nullable
	
	// Relacje
	EventType EventType `gorm:"foreignKey:EventTypeID" json:"event_type,omitempty"`
	Game      Game      `gorm:"foreignKey:GameID" json:"game,omitempty"`
	GamePart  GamePart  `gorm:"foreignKey:GamePartID" json:"game_part,omitempty"`
	Camera    Camera    `gorm:"foreignKey:CameraID" json:"camera,omitempty"`
	Team      *Team     `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Player    *Player   `gorm:"foreignKey:PlayerID" json:"player,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Replay - powtórka
type Replay struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	GameID     uint   `json:"game_id"`
	GamePartID uint   `json:"game_part_id"`
	EventID    uint   `json:"event_id"`
	File       string `json:"file"` // link do pliku
	CameraID   uint   `json:"camera_id"`
	StartTime  int    `json:"start_time"` // czas w sekundach
	EndTime    int    `json:"end_time"`   // czas w sekundach
	
	// Relacje
	Game     Game     `gorm:"foreignKey:GameID" json:"game,omitempty"`
	GamePart GamePart `gorm:"foreignKey:GamePartID" json:"game_part,omitempty"`
	Event    Event    `gorm:"foreignKey:EventID" json:"event,omitempty"`
	Camera   Camera   `gorm:"foreignKey:CameraID" json:"camera,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Substitution - zmiana zawodników
type Substitution struct {
	ID                uint `gorm:"primaryKey" json:"id"`
	GameID            uint `json:"game_id"`
	GamePartID        uint `json:"game_part_id"`
	SubstitutionBlock uint `json:"substitution_block"`
	TeamID            uint `json:"team_id"`
	PlayerInID        uint `json:"player_in_id"`
	PlayerOutID       uint `json:"player_out_id"`
	Time              int  `json:"time"`                          // czas w sekundach
	IsDone            bool `gorm:"default:false" json:"is_done"`
	
	// Relacje
	Game      Game     `gorm:"foreignKey:GameID" json:"game,omitempty"`
	GamePart  GamePart `gorm:"foreignKey:GamePartID" json:"game_part,omitempty"`
	Team      Team     `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	PlayerIn  Player   `gorm:"foreignKey:PlayerInID" json:"player_in,omitempty"`
	PlayerOut Player   `gorm:"foreignKey:PlayerOutID" json:"player_out,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Kit - strój drużyny
type Kit struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	TeamID uint `json:"team_id"`
	Type   int  `gorm:"not null" json:"type"` // 1=home, 2=away, 3=extra
	
	// Relacje
	Team      Team       `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	KitColors []KitColor `gorm:"foreignKey:KitID" json:"kit_colors,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// KitColor - kolor stroju
type KitColor struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	KitID      uint   `json:"kit_id"`
	Color      string `gorm:"not null" json:"color"` // format HEX np. "#FF0000"
	ColorOrder int    `gorm:"not null" json:"color_order"` // kolejność koloru
	
	// Relacje
	Kit Kit `gorm:"foreignKey:KitID" json:"kit,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GamePartValue - wartość statystyki dla drużyny w części meczu
type GamePartValue struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	GameID       uint `json:"game_id"`
	GamePartID   uint `json:"game_part_id"`
	TeamID       uint `json:"team_id"`
	ValueTypeID  uint `json:"value_type_id"`
	Value        int  `json:"value"`                             // wartość statystyki
	IsCumulative bool `gorm:"default:true" json:"is_cumulative"` // czy sumować z poprzednimi częściami
	
	// Relacje
	Game      Game      `gorm:"foreignKey:GameID" json:"game,omitempty"`
	GamePart  GamePart  `gorm:"foreignKey:GamePartID" json:"game_part,omitempty"`
	Team      Team      `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	ValueType ValueType `gorm:"foreignKey:ValueTypeID" json:"value_type,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GameValue - aktualne wartości statystyk dla drużyny w meczu
type GameValue struct {
	ID          uint `gorm:"primaryKey" json:"id"`
	GameID      uint `json:"game_id"`
	TeamID      uint `json:"team_id"`
	ValueTypeID uint `json:"value_type_id"`
	Value       int  `json:"value"` // aktualna obliczona wartość
	
	// Relacje
	Game      Game      `gorm:"foreignKey:GameID" json:"game,omitempty"`
	Team      Team      `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	ValueType ValueType `gorm:"foreignKey:ValueTypeID" json:"value_type,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GamePlayer - zawodnik w meczu
type GamePlayer struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	GameID       uint   `json:"game_id"`
	TeamID       uint   `json:"team_id"`
	PlayerID     uint   `json:"player_id"`
	Number       string `json:"number"` // numer zawodnika w tym meczu
	PlayerRoleID uint   `json:"player_role_id"` // rola/pozycja zawodnika w meczu
	
	// Relacje
	Game       Game       `gorm:"foreignKey:GameID" json:"game,omitempty"`
	Team       Team       `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Player     Player     `gorm:"foreignKey:PlayerID" json:"player,omitempty"`
	PlayerRole PlayerRole `gorm:"foreignKey:PlayerRoleID" json:"player_role,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GameCoach - trener w meczu
type GameCoach struct {
	ID          uint `gorm:"primaryKey" json:"id"`
	GameID      uint `json:"game_id"`
	TeamID      uint `json:"team_id"`
	CoachID     uint `json:"coach_id"`
	CoachRoleID uint `json:"coach_role_id"` // rola trenera w meczu
	
	// Relacje
	Game      Game      `gorm:"foreignKey:GameID" json:"game,omitempty"`
	Team      Team      `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Coach     Coach     `gorm:"foreignKey:CoachID" json:"coach,omitempty"`
	CoachRole CoachRole `gorm:"foreignKey:CoachRoleID" json:"coach_role,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GameReferee - sędzia w meczu
type GameReferee struct {
	ID            uint `gorm:"primaryKey" json:"id"`
	GameID        uint `json:"game_id"`
	RefereeID     uint `json:"referee_id"`
	RefereeRoleID uint `json:"referee_role_id"` // rola sędziego w meczu
	
	// Relacje
	Game        Game        `gorm:"foreignKey:GameID" json:"game,omitempty"`
	Referee     Referee     `gorm:"foreignKey:RefereeID" json:"referee,omitempty"`
	RefereeRole RefereeRole `gorm:"foreignKey:RefereeRoleID" json:"referee_role,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GameTVStaff - członek ekipy TV w meczu
type GameTVStaff struct {
	ID            uint `gorm:"primaryKey" json:"id"`
	GameID        uint `json:"game_id"`
	TVStaffID     uint `json:"tv_staff_id"`
	TVStaffRoleID uint `json:"tv_staff_role_id"` // rola w ekipie realizatorskiej w meczu
	
	// Relacje
	Game        Game        `gorm:"foreignKey:GameID" json:"game,omitempty"`
	TVStaff     TVStaff     `gorm:"foreignKey:TVStaffID" json:"tv_staff,omitempty"`
	TVStaffRole TVStaffRole `gorm:"foreignKey:TVStaffRoleID" json:"tv_staff_role,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Settings - ustawienia aplikacji (singleton - tylko jeden rekord)
type Settings struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Data        string `gorm:"type:text" json:"data"` // JSON przechowywany jako text
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GameCamera - kamera przypisana do meczu
type GameCamera struct {
	GameID   uint `gorm:"primaryKey" json:"game_id"`
	CameraID uint `gorm:"primaryKey" json:"camera_id"`
	IsUsed   bool `gorm:"default:false" json:"is_used"` // czy kamera jest używana w tym meczu
	
	// Relacje
	Game   Game   `gorm:"foreignKey:GameID" json:"game,omitempty"`
	Camera Camera `gorm:"foreignKey:CameraID" json:"camera,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GetAllModels - zwraca slice wszystkich modeli do migracji
func GetAllModels() []interface{} {
	return []interface{}{
		&Team{},
		&PlayerRole{},
		&Player{},
		&CoachRole{},
		&Coach{},
		&RefereeRole{},
		&Referee{},
		&Field{},
		&Competition{},
		&Stage{},
		&Group{},
		&GroupTeam{},
		&Game{},
		&GamePart{},
		&TVStaffRole{},
		&TVStaff{},
		&ValueType{},
		&Camera{},
		&EventType{},
		&Event{},
		&Replay{},
		&Substitution{},
		&Kit{},
		&KitColor{},
		&GamePartValue{},
		&GameValue{},
		&GamePlayer{},
		&GameCoach{},
		&GameReferee{},
		&GameTVStaff{},
		&Settings{},
		&GameCamera{},
	}
}