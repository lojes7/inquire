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

type Conversation struct {
	MyModel
	Type          uint8  `gorm:"smallint;not null"`
	LastMessageID uint64 `gorm:"type:bigint;index"`
}

type MessageUser struct {
	MyModel
	UserID    uint64 `gorm:"bigint;uniqueIndex:idx_message_user"`
	MessageID uint64 `gorm:"bigint;uniqueIndex:idx_message_user"`
	IsDeleted bool   `gorm:"type:boolean;default:false"`
	IsStarred bool   `gorm:"type:boolean;default:false"`
}

type ConversationUser struct {
	MyModel
	UserID         uint64 `gorm:"type:bigint;uniqueIndex:idx_conv_user"`
	ConversationID uint64 `gorm:"type:bigint;uniqueIndex:idx_conv_user"`
	UnreadCount    int    `gorm:"type:int;default:0"`
	IsPinned       bool   `gorm:"type:boolean;default:false"`
	Remark         string `gorm:"varchar(32)"`
}

func (m *Message) BeforeCreate(db *gorm.DB) error {
	m.ID = utils.NewUniqueID()
	return nil
}

func (m *MessageUser) BeforeCreate(db *gorm.DB) error {
	m.ID = utils.NewUniqueID()
	return nil
}

func (c *Conversation) BeforeCreate(db *gorm.DB) error {
	c.ID = utils.NewUniqueID()
	return nil
}

func (c *ConversationUser) BeforeCreate(db *gorm.DB) error {
	c.ID = utils.NewUniqueID()
	return nil
}
