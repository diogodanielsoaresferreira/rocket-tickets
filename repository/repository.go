package repository

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"rocket-tickets/model"
)

var DB *gorm.DB

func Configure() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&model.Event{}, &model.TicketsCategory{}); err != nil {
		return nil, err
	}

	DB = db
	return DB, nil
}

func AddEvent(event *model.Event) error {
	if DB == nil {
		return fmt.Errorf("database not configured")
	}
	return DB.Create(event).Error
}

func GetEvents() ([]model.Event, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not configured")
	}
	var events []model.Event
	err := DB.Preload("Tickets").Find(&events).Error
	return events, err
}

func DeleteEvent(id uint) error {
	if DB == nil {
		return fmt.Errorf("database not configured")
	}
	return DB.Delete(&model.Event{}, id).Error
}
