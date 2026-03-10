package service

import (
	"errors"
	"log"

	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/secure"
	"gorm.io/gorm"
)

// createFriendship 给两个人（传id）双向创建出好友关系
func createFriendship(tx *gorm.DB, userID, friendID uint64) error {
	if tx == nil {
		tx = infra.GetDB()
	}

	var friendName, userName string

	err := tx.Table("users").
		Select("name").
		Where("id = ?", friendID).
		Row().
		Scan(&friendName)

	if err != nil {
		log.Println(err)
		return secure.Wrap(500, "查询好友信息失败", err)
	}

	err = tx.Table("users").
		Select("name").
		Where("id = ?", userID).
		Row().
		Scan(&userName)

	if err != nil {
		log.Println(err)
		return secure.Wrap(500, "查询用户信息失败", err)
	}

	res := tx.Model(&model.Friendship{}).
		Create(model.NewFriendship(userID, friendID, friendName))
	if res.Error != nil {
		log.Println(res.Error)
		return secure.Wrap(500, "创建好友关系失败", res.Error)
	}

	res = tx.Model(&model.Friendship{}).
		Create(model.NewFriendship(friendID, userID, userName))
	if res.Error != nil {
		log.Println(res.Error)
		return secure.Wrap(500, "创建好友关系失败", res.Error)
	}
	return nil
}

func FriendshipList(userID uint64) ([]model.FriendshipListResp, error) {
	var resp []model.FriendshipListResp

	res := infra.GetDB().
		Model(&model.Friendship{}).
		Where("user_id = ?", userID).
		Find(&resp)

	if res.Error != nil {
		log.Println(res.Error)
		return nil, secure.Wrap(500, "加载好友列表失败", res.Error)
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
			return secure.Wrap(500, "删除好友失败", res.Error)
		}
		if res.RowsAffected == 0 {
			log.Println("删除好友操作影响了0行表")
			return secure.Wrap(404, "好友关系不存在", gorm.ErrRecordNotFound)
		}

		res = tx.Where("user_id = ? AND friend_id = ?", friendID, userID).
			Delete(&model.Friendship{})
		if res.Error != nil {
			log.Println(res.Error)
			return secure.Wrap(500, "删除好友失败", res.Error)
		}
		if res.RowsAffected == 0 {
			log.Println("删除好友操作影响了0行表")
			// 这种情况下数据一致性有问题，但还是算作404吧
			return secure.Wrap(404, "好友关系不存在", gorm.ErrRecordNotFound)
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
			return secure.Wrap(500, "修改备注失败", res.Error)
		}
		if res.RowsAffected == 0 {
			log.Println("修改备注操作影响了0行表")
			return secure.Wrap(404, "好友关系不存在", gorm.ErrRecordNotFound)
		}

		// 同时修改会话用户表中的备注
		conversationID, err := getPrivateConversationID(userID, friendID)
		if err != nil {
			return err
		}

		res = tx.Model(&model.ConversationUser{}).
			Where("user_id = ? AND conversation_id = ?", userID, conversationID).
			Update("remark", remark)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			log.Println(res.Error)
			return secure.Wrap(500, "同步会话备注失败", res.Error)
		}
		return nil
	})
}

// isFriend 检查两用户是否为好友关系
// 第一个返回值为true表示是好友关系，false表示不是好友关系
// 需要处理错误
func isFriend(userID, friendID uint64) (bool, error) {
	db := infra.GetDB()

	var cnt int64
	err := db.Model(&model.Friendship{}).
		Where("(user_id = ? AND friend_id = ?)", userID, friendID).
		Count(&cnt).
		Error
	if err != nil {
		log.Println(err)
		return false, secure.Wrap(500, "检查好友关系失败", err)
	}

	if cnt == 0 {
		return false, nil
	}

	return true, nil
}

// getFriendRemark 获取好友备注
func getFriendRemark(tx *gorm.DB, userID, friendID uint64) (string, error) {
	if tx == nil {
		tx = infra.GetDB()
	}

	var remark string
	err := tx.Model(&model.Friendship{}).
		Select("friend_remark").
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Scan(&remark).Error
	if err != nil {
		// 为了统一，辅助函数这里会进行wrap
		return "", secure.Wrap(500, "获取备注失败", err)
	}
	return remark, nil
}
