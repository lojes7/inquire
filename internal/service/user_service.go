package service

import (
	"errors"
	"vvechat/internal/model"
	"vvechat/pkg/secure"

	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db}
}

func (s *UserService) GetUserByUid(uid string) (*model.User, error) {
	var user model.User
	res := s.db.Where("uid = ?", uid).First(&user)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("微信号不存在！")
		}
		return nil, res.Error
	}

	return &user, nil
}

func (s *UserService) IsUidExist(uid string) error {
	var cnt int64
	res := s.db.Model(&model.User{}).Where("uid = ?", uid).Count(&cnt)
	if res.Error != nil {
		return res.Error
	}

	exist := cnt > 0
	if exist {
		return errors.New("Uid is Exist")
	}
	return nil
}

func (s *UserService) IsPhoneNumberExist(phone string) error {
	var cnt int64
	res := s.db.Model(&model.User{}).Where("phone_number = ?", phone).Count(&cnt)
	if res.Error != nil {
		return res.Error
	}

	exist := cnt > 0
	if exist {
		return errors.New("Uid is Exist")
	}
	return nil
}

func (s *UserService) Register(user *model.User) error {
	if err := s.IsPhoneNumberExist(user.PhoneNumber); err != nil {
		return err
	}

	pwd, err := secure.HashString(user.Password)
	if err != nil {
		return err
	}

	user.Password = pwd

	return s.db.Create(user).Error
}

func (s *UserService) LoginByUid(uid string, password string) error {
	user, err := s.GetUserByUid(uid)
	if err != nil {
		return err
	}

	return secure.VerifyPassword(user.Password, password)
}
