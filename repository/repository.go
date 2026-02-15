package repository

import (
	"errors"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"rocket-tickets/model"
)

var ErrEventNotFound = errors.New("event not found")

type EventStore interface {
	AddEvent(event *model.Event) error
	GetEvents() ([]model.Event, error)
	DeleteEvent(id uint) error
}

type GormEventRepository struct {
	db *gorm.DB
}

func NewGormEventRepository(db *gorm.DB) *GormEventRepository {
	return &GormEventRepository{db: db}
}

func Configure() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&model.Event{}, &model.TicketsCategory{}); err != nil {
		return nil, err
	}

	return db, nil
}

func (r *GormEventRepository) AddEvent(event *model.Event) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("database not configured")
	}
	return r.db.Create(event).Error
}

func (r *GormEventRepository) GetEvents() ([]model.Event, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("database not configured")
	}
	var events []model.Event
	err := r.db.Preload("Tickets").Find(&events).Error
	return events, err
}

func (r *GormEventRepository) DeleteEvent(id uint) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("database not configured")
	}

	result := r.db.Delete(&model.Event{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrEventNotFound
	}

	return nil
}
