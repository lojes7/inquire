package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name        string `gorm:"type:varchar(64);not null;uniqueIndex"`
	Password    string `gorm:"type:varchar(72);not null"`
	Uid         string `gorm:"type:varchar(20);not null;uniqueIndex"`
	Region      string `gorm:"type:varchar(32)"`
	PhoneNumber string `gorm:"type:varchar(20);not null;uniqueIndex"`
	Signature   string `gorm:"type:varchar(128);"`
	Gender      string `gorm:"type:varchar(12);"`
}

func NewUser(name string, password string) *User {
	user := User{
		Name:      name,
		Password: password,
	}
	return &user
}

func (*User) TableName() string {
	return "users"
}
