package service

import (
	"errors"
	"log"
	"vvechat/internal/model"
	"vvechat/pkg/infra"

	"gorm.io/gorm"
)

// RefreshToken 刷新Token
func RefreshToken(id uint64) (*model.TokenResp, error) {
	return NewTokenResp(id)
}

func getUserByUid(uid string) (*model.User, error) {
	var user model.User
	res := infra.GetDB().Where("uid = ?", uid).First(&user)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("微信号不存在！")
		}
		log.Println(res.Error)
		return nil, res.Error
	}

	return &user, nil
}

func getUserByPhone(phone string) (*model.User, error) {
	var user model.User
	res := infra.GetDB().Where("phone_number = ?", phone).First(&user)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("手机号不存在！")
		}
		log.Println(res.Error)
		return nil, res.Error
	}

	return &user, nil
}

// 如果数据库查询未出现问题且主键存在返回nil，主键不存在返回invalidData，数据库问题直接返回Error
func isPKExist(id uint64) error {
	var exists int
	err := infra.GetDB().
		Model(&model.User{}).
		Select("1").
		Where("id = ?", id).
		Limit(1).
		Scan(&exists).Error

	if err != nil {
		log.Println(err)
		return err
	}

	// exists == 1 → 存在
	if exists == 1 {
		return nil
	}
	return gorm.ErrInvalidData
}

// createFriendship 给两个人（id主键）创建出好友关系
func createFriendship(tx *gorm.DB, userID, friendID uint64) error {
	var friendName, userName string

	err := tx.Table("users").
		Select("name").
		Where("id = ?", friendID).
		Row().
		Scan(&friendName)

	if err != nil {
		log.Println(err)
		return err
	}

	err = tx.Table("users").
		Select("name").
		Where("id = ?", userID).
		Row().
		Scan(&userName)

	if err != nil {
		log.Println(err)
		return err
	}

	res := tx.Model(&model.Friendship{}).
		Create(model.NewFriendship(userID, friendID, friendName))
	if res.Error != nil {
		log.Println(res.Error)
		return res.Error
	}

	res = tx.Model(&model.Friendship{}).
		Create(model.NewFriendship(friendID, userID, userName))
	if res.Error != nil {
		log.Println(res.Error)
		return res.Error
	}
	return nil
}

// createSystemMessage 在一个会话中创建一个系统级消息
// newID用户指定该系统消息的ID
func createSystemMessage(tx *gorm.DB, content string, conversationID, newID uint64) error {
	newMsg := model.Message{
		SenderID:       0,
		ConversationID: conversationID,
		MyModel: model.MyModel{
			ID: newID,
		},
		Status: model.SYSTEM,
	}
	newText := model.Text{
		Text:      content,
		MessageID: newID,
	}
	res := tx.Create(&newMsg)
	if res.Error != nil {
		log.Println(res.Error)
		return errors.New("创建系统消息失败")
	}

	res = tx.Create(&newText)
	if res.Error != nil {
		log.Println(res.Error)
		return errors.New("创建系统消息失败")
	}
	return nil
}

// updateUnreadCount 给当前会话中除开当前sender的所有人的unread_count加一
func updateUnreadCount(tx *gorm.DB, senderID, conversationID uint64) error {
	res := tx.Model(&model.ConversationUser{}).
		Where("user_id != ? AND conversation_id = ?",
			senderID, conversationID).
		UpdateColumn("unread_count", gorm.Expr("unread_count + ?", 1))
	if res.Error != nil {
		log.Println(res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		log.Println("更新unread count字段影响了0行表")
		return errors.New("更新unread count失败")
	}
	return nil
}

// updateLastMessageID 更新当前会话的last_message_id
func updateLastMessageID(tx *gorm.DB, conversationID, msgID uint64) error {
	res := tx.Model(&model.ConversationUser{}).
		Where("conversation_id = ?", conversationID).
		Update("last_message_id", msgID)
	if res.Error != nil {
		log.Println(res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		log.Println("更新last msg id字段影响了0行表")
		return errors.New("更新last msg id失败")
	}
	return nil
}
