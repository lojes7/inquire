package service

import (
	"errors"
	"vvechat/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db}
}

func (s *UserService) Register(user *model.User) error {
	if user.Uid == "" || user.Password == "" {
		return errors.New("密码或微信号不能为空！")
	}

	var cnt int64
	s.db.Where("id = ?", user.Uid).Count(&cnt)
	if cnt > 0 {
		return errors.New("微信号已存在！")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hash)

	return s.db.Create(user).Error
}
