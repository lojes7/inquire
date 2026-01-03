package service

import (
	"log"
	"vvechat/internal/model"
	"vvechat/pkg/infra"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	return infra.GetDB().
		Model(&model.FriendshipRequest{}).
		Create(model.NewFriendshipRequest(senderID, receiverID, msg, senderName)).
		Error
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
	return infra.GetDB().Transaction(func(tx *gorm.DB) error {

		var req model.FriendshipRequest

		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = ?", id, "pending").
			First(&req).
			Error; err != nil {
			log.Println(err)
			return err
		}

		if err := tx.Model(&req).
			Update("status", "accepted").
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

// FriendRequestReject 拒绝好友申请
func FriendRequestReject(requestID uint64) error {
	return infra.GetDB().
		Model(&model.FriendshipRequest{}).
		Where("id = ? AND status = ?", requestID, "pending").
		Update("status", "rejected").
		Error
}
