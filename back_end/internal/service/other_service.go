package service

import (
	"log"
	"vvechat/internal/model"
	"vvechat/pkg/infra"

	"gorm.io/gorm"
)

// RefreshToken 刷新Token
func RefreshToken(id uint64) (*model.TokenResp, error) {
	return NewTokenResp(id)
}

// FriendInfoByID 查看好友信息
func FriendInfoByID(userID, friendID uint64) (*model.FriendInfoResp, error) {
	var resp model.FriendInfoResp
	resp.ID = friendID
	db := infra.GetDB()

	res := db.Raw(`SELECT f.friend_remark, u.uid, u.name
		FROM friendships f
		JOIN users u 
		ON u.id = f.friend_id
		WHERE f.user_id = ? AND f.friend_id = ?
	`, userID, friendID).Scan(&resp)

	if res.Error != nil {
		log.Println(res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &resp, nil
}

// StrangerInfoByID 查看陌生人信息
func StrangerInfoByID(strangerID uint64) (*model.StrangerInfoResp, error) {
	var resp model.StrangerInfoResp
	resp.ID = strangerID
	db := infra.GetDB()

	res := db.Table("users").
		Select("name").
		Where("id = ?", strangerID).
		First(&resp)
	if res.Error != nil {
		log.Println(res.Error)
		return nil, res.Error
	}

	return &resp, nil
}

// FriendInfoByUid 查看好友信息通过Uid
func FriendInfoByUid(userID uint64, friendUid string) (*model.FriendInfoResp, error) {
	var resp model.FriendInfoResp
	db := infra.GetDB()

	res := db.Raw(`SELECT f.friend_remark, u.uid, u.name, u.id
		FROM friendships f
		JOIN users u 
		ON u.id = f.friend_id
		WHERE f.user_id = ? AND u.uid = ?
	`, userID, friendUid).Scan(&resp)

	if res.Error != nil {
		log.Println(res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &resp, nil
}

// StrangerInfoByUid 查看陌生人信息通过Uid
func StrangerInfoByUid(strangerUid string) (*model.StrangerInfoResp, error) {
	var resp model.StrangerInfoResp
	db := infra.GetDB()

	res := db.Table("users").
		Select("id, name").
		Where("uid = ?", strangerUid).
		First(&resp)
	if res.Error != nil {
		log.Println(res.Error)
		return nil, res.Error
	}

	return &resp, nil
}
