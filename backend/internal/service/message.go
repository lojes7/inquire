package service

import (
	"errors"
	"log"
	"vvechat/internal/model"
	"vvechat/pkg/infra"
	"vvechat/pkg/utils"

	"gorm.io/gorm"
)

func SendMessage(senderID, conversationID uint64, content string) (uint64, error) {
	newID := utils.NewUniqueID()
	newMsg := model.Message{
		SenderID:       senderID,
		ConversationID: conversationID,
		Content:        content,
		MyModel: model.MyModel{
			ID: newID,
		},
	}
	db := infra.GetDB()
	return newID, db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(&newMsg)
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
