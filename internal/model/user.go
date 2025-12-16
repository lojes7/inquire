package model

import (
	"strconv"
	"vvechat/pkg/utils"
)

type User struct {
	ID          uint64 `gorm:"type:bigint;primaryKey;autoIncrement:false"`
	Name        string `gorm:"type:varchar(64);not null;uniqueIndex"`
	Password    string `gorm:"type:varchar(72);not null"`
	Uid         string `gorm:"type:varchar(20);not null;uniqueIndex"`
	Region      string `gorm:"type:varchar(32)"`
	PhoneNumber string `gorm:"type:varchar(20);not null;uniqueIndex"`
	Signature   string `gorm:"type:varchar(128);"`
	Gender      string `gorm:"type:varchar(12);"`
}

func NewUser(name string, password string, phone string) (*User, error) {
	user := User{
		Name:        name,
		Password:    password,
		PhoneNumber: phone,
	}

	id, err := utils.NextUniqueID()
	if err != nil {
		return nil, err
	}

	user.ID = id
	user.Uid = "V_" + strconv.FormatUint(id, 36)
	return &user, nil
}

func (*User) TableName() string {
	return "users"
}
