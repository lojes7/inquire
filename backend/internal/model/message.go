package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"vvechat/pkg/utils"

	"gorm.io/gorm"
)

const (
	TEXT uint8 = iota
	RECALLED
	SYSTEM
	FILE
)

const (
	PRIVATE uint8 = iota
	GROUP
)

type Message struct {
	SenderID       uint64 `gorm:"bigint;index"`
	ConversationID uint64 `gorm:"bigint;index"`
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

type Text struct {
	MyModel
	Text      string `gorm:"varchar(1024);not null"`
	MessageID uint64 `gorm:"bigint;index"`
}

type Vector []float32

type File struct {
	MyModel

	FileName  string `gorm:"type:varchar(255);not null"`
	FileType  string `gorm:"type:varchar(50);not null"`
	FileUrl   string `gorm:"type:varchar(255);not null"`
	FileSize  int64  `gorm:"not null"`
	MessageID uint64 `gorm:"type:bigint;index"`

	FileContent string `gorm:"type:text"`

	ContentVector Vector `gorm:"type:vector(1536)"`
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

func (f *File) BeforeCreate(db *gorm.DB) error {
	if f.ID == 0 {
		f.ID = utils.NewUniqueID()
	}

	return nil
}

func (v Vector) Value() (driver.Value, error) {
	if len(v) == 0 {
		return "[]", nil
	}

	values := make([]string, len(v))
	for i, f := range v {
		values[i] = fmt.Sprintf("%f", f)
	}
	return fmt.Sprintf("[%s]", strings.Join(values, ",")), nil
}

func (v *Vector) Scan(src interface{}) error {
	if src == nil {
		*v = nil
		return nil
	}

	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("cannot scan %T into Vector", src)
	}

	s = strings.Trim(s, "[]")
	if s == "" {
		*v = Vector{}
		return nil
	}

	parts := strings.Split(s, ",")
	vec := make(Vector, len(parts))
	for i, p := range parts {
		fmt.Sscanf(p, "%f", &vec[i])
	}

	*v = vec
	return nil
}
