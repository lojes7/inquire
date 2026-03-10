package service

import (
	"encoding/json"
	"errors"
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
	db := infra.GetDB()

	if senderID == receiverID {
		return secure.Wrap(400, "不能给自己发送好友请求", gorm.ErrInvalidData)
	}
	if err := isPKExist(receiverID); err != nil {
		log.Println(err)
		return secure.Wrap(400, "接收者不存在", err)
	}

	var count int64
	db.Model(&model.Friendship{}).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			senderID, receiverID, receiverID, senderID).
		Count(&count)
	if count > 0 {
		return secure.Wrap(400, "对方已经是你的好友", gorm.ErrDuplicatedKey)
	}

	// 检查是否有待处理的申请，防止重复发送
	var requestCount int64
	db.Model(&model.FriendshipRequest{}).
		Where("sender_id = ? AND receiver_id = ? AND status = ?", senderID, receiverID, PENDING).
		Count(&requestCount)
	if requestCount > 0 {
		return secure.Wrap(400, "好友申请已发送，请耐心等待", gorm.ErrDuplicatedKey)
	}

	err := db.Model(&model.FriendshipRequest{}).
		Create(model.NewFriendshipRequest(senderID, receiverID, msg, senderName)).
		Error

	if err != nil {
		log.Println(err)
		return secure.Wrap(500, "发送好友申请失败", err)
	}

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
	// 异步发送通知，避免阻塞主流程（视ws实现而定，如果SendToUser是非阻塞的可以直接调用）
	ws.GetHub().SendToUser(receiverID, msgBytes)

	return nil
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
		return nil, secure.Wrap(500, "加载申请列表失败", res.Error)
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
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return secure.Wrap(404, "申请不存在或已处理", err)
			}
			return secure.Wrap(500, "处理申请失败", err)
		}

		if err := tx.Model(&req).
			Update("status", ACCEPTED).
			Error; err != nil {
			log.Println(err)
			return secure.Wrap(500, "更新申请状态失败", err)
		}

		err := createFriendship(tx, req.SenderID, req.ReceiverID)
		if err != nil {
			log.Println(err)
			return secure.Wrap(500, "创建好友关系失败", err)
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
		return secure.Wrap(500, "删除申请失败", res.Error)
	}
	if res.RowsAffected == 0 {
		log.Println("删除好友申请操作影响了0行表")
		return secure.Wrap(404, "申请不存在", gorm.ErrRecordNotFound)
	}
	return nil
}
