package model

import (
	"strconv"

	"github.com/lojes7/inquire/pkg/utils"
	"gorm.io/gorm"
)

type User struct {
	MyModel
	Name        string `gorm:"type:varchar(64);not null"`
	Password    string `gorm:"type:varchar(72);not null"`
	Uid         string `gorm:"type:varchar(20);not null;uniqueIndex"`
	Region      string `gorm:"type:varchar(32)"`
	PhoneNumber string `gorm:"type:varchar(20);not null;uniqueIndex"`
	Signature   string `gorm:"type:varchar(128);"`
	Gender      string `gorm:"type:varchar(12);check:gender IN ('male','female','')"`
}

func NewUser(name string, password string, phone string) (*User, error) {
	user := User{
		Name:        name,
		Password:    password,
		PhoneNumber: phone,
	}

	return &user, nil
}

func (*User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(db *gorm.DB) error {
	u.ID = utils.NewUniqueID()
	u.Uid = "V_" + strconv.FormatUint(u.ID, 36)
	return nil
}
