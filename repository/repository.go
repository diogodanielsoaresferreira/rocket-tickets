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
	GetEvent(id uint) (model.Event, error)
	UpdateEvent(id uint, updatedEvent *model.Event) error
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

func (r *GormEventRepository) GetEvent(id uint) (model.Event, error) {
	if r == nil || r.db == nil {
		return model.Event{}, fmt.Errorf("database not configured")
	}

	var event model.Event
	err := r.db.Preload("Tickets").First(&event, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Event{}, ErrEventNotFound
	}
	return event, err
}

func (r *GormEventRepository) UpdateEvent(id uint, updatedEvent *model.Event) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("database not configured")
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var existingEvent model.Event
		err := tx.First(&existingEvent, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEventNotFound
		}
		if err != nil {
			return err
		}

		updatedEvent.ID = existingEvent.ID

		// Use an explicit map so event zero values are included in SQL updates.
		if err := tx.Model(&model.Event{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"title":  updatedEvent.Title,
				"date":   updatedEvent.Date,
				"venue":  updatedEvent.Venue,
				"artist": updatedEvent.Artist,
			}).Error; err != nil {
			return err
		}

		if err := tx.Where("event_id = ?", id).Delete(&model.TicketsCategory{}).Error; err != nil {
			return err
		}

		if updatedEvent.Tickets == nil || len(*updatedEvent.Tickets) == 0 {
			return nil
		}

		tickets := *updatedEvent.Tickets
		for i := range tickets {
			tickets[i].ID = 0
			tickets[i].EventID = id
		}

		if err := tx.Create(&tickets).Error; err != nil {
			return err
		}

		updatedEvent.Tickets = &tickets
		return nil
	})
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
