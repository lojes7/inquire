package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	TEXT int = iota
	RECALLED
	SYSTEM
	FILE
)

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
