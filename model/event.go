package model

import (
	"time"

	"gorm.io/gorm"
)

type TicketsCategory struct {
	gorm.Model
	EventID   uint    `json:"-"`
	Category  string  `json:"category" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	Available int     `json:"available" binding:"required"`
}

type Event struct {
	gorm.Model
	Title   string             `json:"title" binding:"required"`
	Date    time.Time          `json:"date" binding:"required"`
	Venue   string             `json:"venue" binding:"required"`
	Artist  string             `json:"artist" binding:"required"`
	Tickets *[]TicketsCategory `json:"tickets" gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE;" binding:"omitempty,dive"`
}
