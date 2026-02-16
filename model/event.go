package model

import (
	"time"

	"gorm.io/gorm"
)

type TicketsCategory struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	EventID   uint           `json:"-"`
	Category  string         `json:"category" binding:"required"`
	Price     float64        `json:"price" binding:"gte=0"`
	Quantity  int            `json:"quantity" binding:"required,gt=0"`
	Available int            `json:"available" binding:"gte=0,ltefield=Quantity"`
}

type Event struct {
	ID        uint               `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt     `json:"-" gorm:"index"`
	Title     string             `json:"title" binding:"required"`
	Date      time.Time          `json:"date" binding:"required"`
	Venue     string             `json:"venue" binding:"required"`
	Artist    string             `json:"artist" binding:"required"`
	Tickets   *[]TicketsCategory `json:"tickets" gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE;" binding:"omitempty,dive"`
}

type Ticket struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	TicketCategoryID uint      `json:"-"`
	Status           string    `json:"status" gorm:"type:text;default:sold"`
	SoldAt           time.Time `json:"soldAt,omitempty"`
}
