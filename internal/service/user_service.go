package service

import (
	"errors"
	"vvechat/pkg/infra"
	"vvechat/internal/model"
	"vvechat/pkg/secure"

	"gorm.io/gorm"
)

func GetUserByUid(uid string) (*model.User, error) {
	var user model.User
	res := infra.GetDB().Where("uid = ?", uid).First(&user)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("微信号不存在！")
		}
		return nil, res.Error
	}

	return &user, nil
}

func GetUserByPhone(phone string) (*model.User, error) {
	var user model.User
	res := infra.GetDB().Where("phone_number = ?", phone).First(&user)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("手机号不存在！")
		}
		return nil, res.Error
	}

	return &user, nil
}

func IsUidExist(uid string) error {
	var cnt int64
	res := infra.GetDB().Model(&model.User{}).Where("uid = ?", uid).Count(&cnt)
	if res.Error != nil {
		return res.Error
	}

	exist := cnt > 0
	if exist {
		return errors.New("微信号重复！")
	}
	return nil
}

func IsPhoneNumberExist(phone string) error {
	var cnt int64
	res := infra.GetDB().Model(&model.User{}).Where("phone_number = ?", phone).Count(&cnt)
	if res.Error != nil {
		return res.Error
	}

	exist := cnt > 0
	if exist {
		return errors.New("该手机号已被注册！")
	}
	return nil
}

func Register(user *model.User) error {
	if err := IsPhoneNumberExist(user.PhoneNumber); err != nil {
		return err
	}

	pwd, err := secure.HashString(user.Password)
	if err != nil {
		return err
	}

	user.Password = pwd

	return infra.GetDB().Create(user).Error
}

func LoginByUid(uid string, password string) error {
	user, err := GetUserByUid(uid)
	if err != nil {
		return err
	}

	return secure.VerifyPassword(user.Password, password)
}

func LoginByPhone(phone string, password string) error {
	user, err := GetUserByPhone(phone)
	if err != nil {
		return err
	}

	return secure.VerifyPassword(user.Password, password)
}
