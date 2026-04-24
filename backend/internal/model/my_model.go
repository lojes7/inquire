package model

import (
	"time"

	"gorm.io/gorm"
)

// 消息状态常量
const (
	TEXT int = iota
	RECALLED
	SYSTEM
	FILE
)

// 会话类型常量
const (
	PRIVATE int = iota
	GROUP
)

type MyModel struct {
	ID        uint64 `gorm:"primaryKey;type:bigint"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAT gorm.DeletedAt
}
