package model

import (
	"github.com/lojes7/inquire/pkg/utils"
	"gorm.io/gorm"
)

type Message struct {
	SenderID       uint64 `gorm:"bigint;index"`
	ConversationID uint64 `gorm:"bigint;index"`
	Status         int    `gorm:"smallint;default:0"`
	FileID         uint64 `gorm:"bigint;index"`
	Content        string `gorm:"varchar(1024);not null"`
	MyModel
}

type MessageUser struct {
	MyModel
	UserID    uint64 `gorm:"bigint;uniqueIndex:idx_message_user"`
	MessageID uint64 `gorm:"bigint;uniqueIndex:idx_message_user"`
	IsStarred bool   `gorm:"type:boolean;default:false"`
	IsDeleted bool   `gorm:"type:boolean;default:false"`
}

type Conversation struct {
	MyModel
	Type      int    `gorm:"smallint;not null"`
	OwnerID   uint64 `gorm:"bigint;index"`
	GroupName string `gorm:"varchar(64)"`
}

type ConversationUser struct {
	MyModel
	UserID         uint64 `gorm:"type:bigint;uniqueIndex:idx_conv_user"`
	ConversationID uint64 `gorm:"type:bigint;uniqueIndex:idx_conv_user"`
	UnreadCount    int    `gorm:"type:int;default:0"`
	Remark         string `gorm:"varchar(32)"`
	LastMessageID  uint64 `gorm:"type:bigint;index"`
	IsPinned       bool   `gorm:"type:boolean;default:false"`
	IsBanned       bool   `gorm:"type:boolean;default:false"`
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
