package model

import (
	"strconv"
	"time"
	"vvechat/pkg/utils"

	"gorm.io/gorm"
)

type User struct {
	ID          uint64         `gorm:"type:bigint;primaryKey;autoIncrement:false"`
	Name        string         `gorm:"type:varchar(64);not null;uniqueIndex"`
	Password    string         `gorm:"type:varchar(72);not null"`
	Uid         string         `gorm:"type:varchar(20);not null;uniqueIndex"`
	Region      string         `gorm:"type:varchar(32)"`
	PhoneNumber string         `gorm:"type:varchar(20);not null;uniqueIndex"`
	Signature   string         `gorm:"type:varchar(128);"`
	Gender      string         `gorm:"type:varchar(12);check:gender IN ('male','female','')"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	CreatedAt   time.Time      `gorm:"not null;autoCreateTime"`
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
