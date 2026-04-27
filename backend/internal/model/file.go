package model

import (
	"github.com/lojes7/inquire/pkg/utils"
	"gorm.io/gorm"
)

type File struct {
	MyModel

	FileName  string `gorm:"type:varchar(255);not null"`
	FileType  string `gorm:"type:varchar(50);not null"`
	FileURL   string `gorm:"type:varchar(255);not null"`
	FileSize  int64  `gorm:"not null"`
	HashValue string `gorm:"type:char(64);not null;"`
}

type UserFile struct {
	MyModel
	UserID uint64 `gorm:"bigint;index;not null"`
	FileID uint64 `gorm:"bigint;index;not null"`
}

func (f *File) BeforeCreate(db *gorm.DB) error {
	if f.ID == 0 {
		f.ID = utils.NewUniqueID()
	}

	return nil
}

func (uf *UserFile) BeforeCreate(db *gorm.DB) error {
	if uf.ID == 0 {
		uf.ID = utils.NewUniqueID()
	}

	return nil
}
