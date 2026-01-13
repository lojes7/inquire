package model

import (
	"vvechat/pkg/utils"

	"gorm.io/gorm"
)

const (
	DEFAULT uint8 = iota
	RECALLED
	SYSTEM
)

const (
	PRIVATE uint8 = iota
	GROUP
)

type Message struct {
	SenderID       uint64 `gorm:"bigint;index"`
	ConversationID uint64 `gorm:"bigint;index"`
	Content        string `gorm:"varchar(1024);not null"`
	Status         uint8  `gorm:"smallint;default:0"`
	MyModel
}

type MessageUser struct {
	MyModel
	UserID    uint64 `gorm:"bigint;uniqueIndex:idx_message_user"`
	MessageID uint64 `gorm:"bigint;uniqueIndex:idx_message_user"`
	IsDeleted bool   `gorm:"type:boolean;default:false"`
	IsStarred bool   `gorm:"type:boolean;default:false"`
}

type Conversation struct {
	MyModel
	// 如果是私聊（type为0）则conversation_id 是friend_id
	// 如果是群里（type为1）则conversation_id 由雪花ID生成器分配
	Type uint8 `gorm:"smallint;not null"`
}
type ConversationUser struct {
	MyModel
	UserID         uint64 `gorm:"type:bigint;uniqueIndex:idx_conv_user"`
	ConversationID uint64 `gorm:"type:bigint;uniqueIndex:idx_conv_user"`
	UnreadCount    int    `gorm:"type:int;default:0"`
	Remark         string `gorm:"varchar(32)"`
	LastMessageID  uint64 `gorm:"type:bigint;index"`
	IsDeleted      bool   `gorm:"type:boolean;default:false"`
	IsPinned       bool   `gorm:"type:boolean;default:false"`
}

func (m *Message) BeforeCreate(db *gorm.DB) error {
	if m.ID == 0 {
		m.ID = utils.NewUniqueID()
	}
	return nil
}

func (m *MessageUser) BeforeCreate(db *gorm.DB) error {
	if m.ID == 0 {
		m.ID = utils.NewUniqueID()
	}
	return nil
}

func (c *Conversation) BeforeCreate(db *gorm.DB) error {
	if c.ID == 0 {
		c.ID = utils.NewUniqueID()
	}
	return nil
}

func (c *ConversationUser) BeforeCreate(db *gorm.DB) error {
	if c.ID == 0 {
		c.ID = utils.NewUniqueID()
	}
	return nil
}
