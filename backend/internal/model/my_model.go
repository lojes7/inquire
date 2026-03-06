package model

import (
	"time"
)

type MyModel struct {
	ID        uint64 `gorm:"primaryKey;type:bigint"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
