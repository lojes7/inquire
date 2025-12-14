package class

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name          string `gorm:"not null;unique;size:64"`
	Password      string `gorm:"not null"`
	Uid           string `gorm:"unique"`
	Email         string `gorm:"unique"`
	ClientIP      string
	ClientPort    string
	PhoneNumber   string    `gorm:"unique"`
	LoginTime     time.Time `gorm:"default:now()"`
	LoginOutTime  time.Time
	HeartbeatTime time.Time
	IsLogOut      bool
	DeviceInfo    string
}

func NewUser(uid string, password string) *User {
	user := User{
		Uid: uid,
		Password: password,
	}
	return &user
}

func AddUserToDB(user *User, db *gorm.DB) *gorm.DB {
	return db.Create(user)
}