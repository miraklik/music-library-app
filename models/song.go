package models

import (
	"time"
)

type Group struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	Name      string     `json:"name" gorm:"unique;not null"`
	Songs     []Song     `json:"songs"`
}

type Song struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	GroupID     uint       `json:"group_id" gorm:"index"`
	Group       Group      `json:"group"`
	Song        string     `json:"song" gorm"index"`
	ReleaseDate time.Time  `json:"release_date" gorm:"index"`
	Text        string     `json:"text"`
	Link        string     `json:"link"`
}

type SongDetail struct {
	ReleaseDate string `json:"release_date"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type NewSongRequest struct {
	Group string `json:"group" binding:"required"`
	Song  string `json:"song" binding:"required"`
}
