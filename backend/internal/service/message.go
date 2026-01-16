package service

import (
	"errors"
	"log"
	"vvechat/internal/model"
	"vvechat/pkg/infra"
	"vvechat/pkg/utils"

	"gorm.io/gorm"
)

func SendText(senderID, conversationID uint64, content string) (uint64, error) {
	newID := utils.NewUniqueID()
	newMsg := model.Message{
		SenderID:       senderID,
		ConversationID: conversationID,
		Status:         model.TEXT,
		MyModel: model.MyModel{
			ID: newID,
		},
	}
	newText := model.Text{
		Text:      content,
		MessageID: newID,
	}
	db := infra.GetDB()
	return newID, db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(&newMsg)
		if res.Error != nil {
			log.Println(res.Error)
			return errors.New("服务器错误")
		}

		res = tx.Create(&newText)
		if res.Error != nil {
			log.Println(res.Error)
			return errors.New("服务器错误")
		}

		err := updateLastMessageID(tx, conversationID, newID)
		if err != nil {
			return errors.New("服务器错误")
		}

		err = updateUnreadCount(tx, senderID, conversationID)
		if err != nil {
			return errors.New("服务器错误")
		}

		return nil
	})
}

func RecallMessage(userID, msgID uint64) (uint64, error) {
	db := infra.GetDB()
	var temp model.Message
	err := db.Model(&model.Message{}).
		Select("sender_id, conversation_id").
		Where("id = ?", msgID).
		First(&temp).
		Error
	if err != nil {
		log.Println(err)
		return 0, errors.New("服务器错误")
	}

	senderID := temp.SenderID
	conversationID := temp.ConversationID
	if senderID != userID {
		return 0, errors.New("不能撤回不是自己发的消息")
	}

	newID := utils.NewUniqueID()
	return newID, db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Message{}).
			Where("id = ?", msgID).
			Update("status", model.RECALLED)
		if res.Error != nil {
			log.Println(res.Error)
			return errors.New("服务器错误")
		}
		if res.RowsAffected == 0 {
			log.Println("撤回消息操作影响了0行表")
			return errors.New("服务器错误")
		}

		var senderName string
		err = tx.Model(&model.User{}).
			Where("id = ?", senderID).
			Pluck("name", &senderName).Error
		if err != nil {
			log.Println(err)
			return errors.New("服务器错误")
		}
		newContent := senderName + "撤回了一条消息"
		err = createSystemMessage(tx, newContent, conversationID, newID)
		if err != nil {
			return err
		}

		err = updateLastMessageID(tx, conversationID, newID)
		if err != nil {
			return err
		}
		return nil
	})
}

func DeleteMessage(userID, messageID uint64) error {
	db := infra.GetDB()
	var temp model.Message
	err := db.Model(&model.Message{}).
		Select("conversation_id").
		Where("id = ?", messageID).
		First(&temp).Error
	if err != nil {
		log.Println(err)
		return errors.New("服务器错误")
	}
	conversationID := temp.ConversationID

	return db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.MessageUser{}).
			Where("message_id = ? AND user_id = ?", messageID, userID).
			Update("is_deleted", true)
		if res.Error != nil {
			log.Println(res.Error)
			return errors.New("服务器错误")
		}
		if res.RowsAffected == 0 {
			log.Println("删除消息操作影响了0行表")
			return errors.New("服务器错误")
		}

		var lastID uint64
		sql := `SELECT m.id FROM messages m 
			LEFT JOIN message_users mu ON mu.message_id = m.id AND mu.user_id = ?
			WHERE m.status != ? AND mu.is_deleted = false
			ORDER BY m.created_at DESC 
			LIMIT 1`
		res = tx.Raw(sql, userID, model.RECALLED).Scan(&lastID)
		if res.Error != nil {
			log.Println(res.Error)
			return errors.New("服务器错误")
		}
		if res.RowsAffected == 0 {
			log.Println("删除消息更新最后消息id 时没有查到id")
			return errors.New("服务器错误")
		}

		err = updateLastMessageID(tx, conversationID, lastID)
		if err != nil {
			return err
		}

		return nil
	})
}
