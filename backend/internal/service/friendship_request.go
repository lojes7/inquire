package service

import (
	"encoding/json"
	"log"

	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/internal/ws"
	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/secure"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ACCEPTED = "accepted"
	PENDING  = "pending"
	REJECTED = "rejected"
	CANCELED = "canceled"
)

// SendFriendRequest 发送好友申请操作
func SendFriendRequest(senderID, receiverID uint64, msg string, senderName string) error {
	if senderID == receiverID {
		return gorm.ErrInvalidData
	}
	if err := isPKExist(receiverID); err != nil {
		log.Println(err)
		return err
	}

	var count int64
	infra.GetDB().
		Model(&model.Friendship{}).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			senderID, receiverID, receiverID, senderID).
		Count(&count)
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}

	err := infra.GetDB().
		Model(&model.FriendshipRequest{}).
		Create(model.NewFriendshipRequest(senderID, receiverID, msg, senderName)).
		Error

	if err == nil {
		// 发送 websocket 通知
		notification := map[string]any{
			"type": "new_friend_request",
			"data": map[string]any{
				"sender_id":   senderID,
				"sender_name": senderName,
				"message":     msg,
			},
		}
		msgBytes, _ := json.Marshal(notification)
		ws.GetHub().SendToUser(receiverID, msgBytes)
	}

	return &secure.MyError{
		Err:     err,
		Message: "发送好友申请失败",
		Code:    500,
	}
}

// FriendRequestList 加载好友申请列表操作
func FriendRequestList(receiverID uint64) ([]model.FriendRequestListResp, error) {
	respSlice := make([]model.FriendRequestListResp, 0)

	res := infra.GetDB().
		Model(&model.FriendshipRequest{}).
		Where("receiver_id = ?", receiverID).
		Order("created_at DESC").
		Find(&respSlice)
	if res.Error != nil {
		log.Println(res.Error)
		return nil, res.Error
	}

	return respSlice, nil
}

// FriendRequestAccept 通过好友申请
func FriendRequestAccept(id uint64) error {
	// 参数的 id 是好友申请表的主键 ID
	return infra.GetDB().Transaction(func(tx *gorm.DB) error {

		var req model.FriendshipRequest

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = ?", id, PENDING).
			First(&req).
			Error; err != nil {
			log.Println(err)
			return err
		}

		if err := tx.Model(&req).
			Update("status", ACCEPTED).
			Error; err != nil {
			log.Println(err)
			return err
		}

		err := createFriendship(tx, req.SenderID, req.ReceiverID)
		if err != nil {
			log.Println(err)
			return err
		}

		return nil
	})
}

// FriendRequestDelete 删除好友申请
func FriendRequestDelete(requestID uint64) error {
	db := infra.GetDB()

	var req model.FriendshipRequest
	req.ID = requestID

	res := db.Delete(&req)
	if res.Error != nil {
		log.Println(res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		log.Println("删除好友申请操作影响了0行表")
		return gorm.ErrRecordNotFound
	}
	return nil
}
