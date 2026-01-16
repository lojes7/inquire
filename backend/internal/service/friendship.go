package service

import (
	"errors"
	"log"
	"vvechat/internal/model"
	"vvechat/pkg/infra"

	"gorm.io/gorm"
)

func FriendshipList(userID uint64) ([]model.FriendshipListResp, error) {
	var resp []model.FriendshipListResp

	res := infra.GetDB().
		Model(&model.Friendship{}).
		Where("user_id = ?", userID).
		Find(&resp)

	if res.Error != nil {
		log.Println(res.Error)
		return nil, res.Error
	}

	return resp, nil
}

func DeleteFriendship(userID, friendID uint64) error {
	db := infra.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("user_id = ? AND friend_id = ?", userID, friendID).
			Delete(&model.Friendship{})
		if res.Error != nil {
			log.Println(res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			log.Println("删除好友操作影响了0行表")
			return gorm.ErrRecordNotFound
		}

		res = tx.Where("user_id = ? AND friend_id = ?", friendID, userID).
			Delete(&model.Friendship{})
		if res.Error != nil {
			log.Println(res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			log.Println("删除好友操作影响了0行表")
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

func ReviseRemark(userID, friendID uint64, remark string) error {
	db := infra.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Friendship{}).
			Where("user_id = ? AND friend_id = ?", userID, friendID).
			Update("friend_remark", remark)
		if res.Error != nil {
			log.Println(res.Error)
			return errors.New("服务器错误")
		}
		if res.RowsAffected == 0 {
			log.Println("修改备注操作影响了0行表")
			return errors.New("服务器错误")
		}

		res = tx.Model(&model.ConversationUser{}).
			Where("user_id = ? AND conversation_id = ?", userID, friendID).
			Update("remark", remark)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			log.Println(res.Error)
			return errors.New("服务器错误")
		}
		return nil
	})
}
