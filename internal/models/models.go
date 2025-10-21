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

// Game - mecz
type Game struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	FieldID  uint   `json:"field_id"`
	DateTime string `json:"date_time"` // Format: "2025-10-17_20:45"
	
	// Relacje
	Field Field `gorm:"foreignKey:FieldID" json:"field,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Gamepart - część meczu (kwarta, połowa, etc.)
type Gamepart struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	Name       string  `gorm:"not null" json:"name"`
	MinPoints  *int    `json:"min_points"` // nullable
	Duration   *int    `json:"duration"`   // nullable - czas w sekundach
	MatchOrder int     `gorm:"unique;not null" json:"match_order"`
	
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

// Eventtype - typ wydarzenia
type Eventtype struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	Name         string  `gorm:"not null" json:"name"`
	Icon         *string `json:"icon"` // nullable - URL do ikonki
	ShortName    string  `json:"short_name"`
	IsInProtocol bool    `gorm:"default:false" json:"is_in_protocol"`
	
	// Relacje
	Events []Event `gorm:"foreignKey:EventtypeID" json:"events,omitempty"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Event - wydarzenie w meczu
type Event struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	EventtypeID uint   `json:"eventtype_id"`
	GameID      uint   `json:"game_id"`
	GamepartID  uint   `json:"gamepart_id"`
	Eventtime   int    `json:"eventtime"` // czas w sekundach
	CameraID    uint   `json:"camera_id"`
	TeamID      *uint  `json:"team_id"`   // nullable
	PlayerID    *uint  `json:"player_id"` // nullable
	
	// Relacje
	Eventtype Eventtype `gorm:"foreignKey:EventtypeID" json:"eventtype,omitempty"`
	Game      Game      `gorm:"foreignKey:GameID" json:"game,omitempty"`
	Gamepart  Gamepart  `gorm:"foreignKey:GamepartID" json:"gamepart,omitempty"`
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
	GamepartID uint   `json:"gamepart_id"`
	EventID    uint   `json:"event_id"`
	File       string `json:"file"` // link do pliku
	CameraID   uint   `json:"camera_id"`
	Starttime  int    `json:"starttime"` // czas w sekundach
	Endtime    int    `json:"endtime"`   // czas w sekundach
	
	// Relacje
	Game     Game     `gorm:"foreignKey:GameID" json:"game,omitempty"`
	Gamepart Gamepart `gorm:"foreignKey:GamepartID" json:"gamepart,omitempty"`
	Event    Event    `gorm:"foreignKey:EventID" json:"event,omitempty"`
	Camera   Camera   `gorm:"foreignKey:CameraID" json:"camera,omitempty"`
	
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
		&Game{},
		&Gamepart{},
		&TVStaffRole{},
		&TVStaff{},
		&ValueType{},
		&Camera{},
		&Eventtype{},
		&Event{},
		&Replay{},
	}
}