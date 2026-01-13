package model

import (
	"vvechat/pkg/utils"

	"gorm.io/gorm"
)

// Friendship 好友关系表
// userID、friendID、deleted_at为联合唯一索引
type Friendship struct {
	MyModel
	UserID       uint64 `gorm:"type:bigint;not null"`
	FriendID     uint64 `gorm:"type:bigint;not null"`
	FriendRemark string `gorm:"type:varchar(64);not null"`
}

// FriendshipRequest 好友申请列表
// senderID、receiverID、deleted_at为联合唯一索引
type FriendshipRequest struct {
	MyModel
	SenderID            uint64 `gorm:"type:bigint;not null;index:idx_sender_receiver,unique"`
	SenderName          string `gorm:"type:varchar(64);not null;"`
	ReceiverID          uint64 `gorm:"type:bigint;not null;index:idx_sender_receiver,unique"`
	VerificationMessage string `gorm:"type:varchar(128)"`
	Status              string `gorm:"type:varchar(16);not null;check:status IN ('pending','accepted','rejected','canceled')"`
}

func NewFriendship(userID, friendID uint64, remark string) *Friendship {
	return &Friendship{
		UserID:       userID,
		FriendID:     friendID,
		FriendRemark: remark,
	}
}

func NewFriendshipRequest(senderID, receiverID uint64, msg string, senderName string) *FriendshipRequest {
	return &FriendshipRequest{
		SenderID:            senderID,
		SenderName:          senderName,
		ReceiverID:          receiverID,
		VerificationMessage: msg,
		Status:              "pending",
	}
}

func (f *FriendshipRequest) BeforeCreate(db *gorm.DB) error {
	if f.ID == 0 {
		f.ID = utils.NewUniqueID()
	}
	return nil
}

func (f *Friendship) BeforeCreate(db *gorm.DB) error {
	if f.ID == 0 {
		f.ID = utils.NewUniqueID()
	}

	return nil
}
