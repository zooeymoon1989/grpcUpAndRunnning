package models

import (
	"time"
)

type MyGuest struct {
	Id        int64     `gorm:"id;primaryKey"`
	Firstname string    `gorm:"firstname"`
	Lastname  string    `gorm:"lastname"`
	Email     string    `gorm:"email"`
	RegDate   time.Time `gorm:"reg_date"`
}

func (m *MyGuest) TableName() string {
	return "MyGuests"
}
