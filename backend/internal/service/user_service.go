package service

import (
	"errors"
	"log"
	"vvechat/internal/model"
	"vvechat/pkg/infra"
	"vvechat/pkg/secure"

	"gorm.io/gorm"
)

func NewTokenResp(id uint64) (*model.TokenResp, error) {
	var resp model.TokenResp

	token, err := secure.NewToken(id)
	if err != nil {
		log.Println(err)
		return nil, errors.New("token生成错误" + err.Error())
	}
	refreshToken, err := secure.NewRefreshToken(id)
	if err != nil {
		log.Println(err)
		return nil, errors.New("refreshToken生成错误" + err.Error())
	}

	t := uint64(secure.GetExpiresTime().Seconds())
	if t <= 0 {
		log.Println("生成token时viper解析失败")
		return nil, errors.New("生成token时viper解析失败")
	}
	resp.ExpiresIn = t
	resp.Token, resp.RefreshToken = token, refreshToken

	return &resp, nil
}

func NewLoginResp(name string, uid string, id uint64) (*model.LoginResp, error) {
	var resp model.LoginResp

	tokenClass, err := NewTokenResp(id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	resp.TokenClass.ExpiresIn = tokenClass.ExpiresIn
	resp.TokenClass.Token = tokenClass.Token
	resp.TokenClass.RefreshToken = tokenClass.RefreshToken

	resp.UserInfo.Name, resp.UserInfo.Uid = name, uid

	return &resp, nil
}

// Register 注册操作
func Register(user *model.User) error {
	pwd, err := secure.HashString(user.Password)
	if err != nil {
		log.Println(err)
		return err
	}

	user.Password = pwd

	return infra.GetDB().Create(user).Error
}

// LoginByUid 微信号登陆操作
func LoginByUid(uid string, password string) (*model.LoginResp, error) {
	user, err := getUserByUid(uid)
	if err != nil {
		log.Println(err)
		return nil, errors.New("登陆失败 微信号或密码错误")
	}

	if err := secure.VerifyPassword(user.Password, password); err != nil {
		log.Println(err)
		return nil, errors.New("登陆失败 微信号或密码错误")
	}

	return NewLoginResp(user.Name, user.Uid, user.ID)
}

// LoginByPhone 手机号登陆操作
func LoginByPhone(phone string, password string) (*model.LoginResp, error) {
	user, err := getUserByPhone(phone)
	if err != nil {
		log.Println(err)
		return nil, errors.New("登陆失败 手机号或密码错误")
	}

	if err := secure.VerifyPassword(user.Password, password); err != nil {
		log.Println(err)
		return nil, errors.New("登陆失败 手机号或密码错误")
	}

	return NewLoginResp(user.Name, user.Uid, user.ID)
}

// ReviseUid 修改微信号
func ReviseUid(id uint64, newUid string) error {
	return infra.GetDB().
		Model(&model.User{}).
		Where("id = ?", id).
		Update("uid", newUid).
		Error
}

// RevisePassword 修改密码
func RevisePassword(id uint64, prevPassword, newPassword string) error {
	if prevPassword == newPassword {
		log.Println("修改密码时传入了相同的密码")
		return errors.New("新密码与旧密码不能相同")
	}

	db := infra.GetDB()
	var user model.User

	res := db.Select("id, password").
		Where("id = ?", id).
		First(&user)

	if res.Error != nil {
		log.Println(res.Error)
		return errors.New("服务器错误")
	}
	if user.Password == "" {
		log.Println("修改密码时查询数据库获取哈希密码失败")
		return errors.New("服务器错误")
	}

	err := secure.VerifyPassword(user.Password, prevPassword)
	if err != nil {
		log.Println(err)
		return errors.New("密码错误！")
	}

	newHashPassword, err := secure.HashString(newPassword)
	if err != nil {
		log.Println(err)
		return errors.New("服务器错误")
	}

	res = db.Model(&user).
		Update("password", newHashPassword)
	if res.Error != nil {
		log.Println(res.Error)
		return errors.New("服务器错误")
	}
	if res.RowsAffected == 0 {
		log.Println("修改密码操作影响了0行表")
		return errors.New("服务器错误")
	}

	return nil
}

// ReviseName 修改用户名
func ReviseName(id uint64, newName string) error {
	return infra.GetDB().Model(model.User{}).
		Where("id = ?", id).
		Update("name", newName).Error
}
