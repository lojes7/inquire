package model

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/lojes7/inquire/pkg/utils"
	"gorm.io/gorm"
)

type File struct {
	MyModel

	FileName      string `gorm:"type:varchar(255);not null"`
	FileType      string `gorm:"type:varchar(50);not null"`
	FileURL       string `gorm:"type:varchar(255);not null"`
	FileSize      int64  `gorm:"not null"`
	HashValue     string `gorm:"type:char(64);not null;"`
	ContentVector Vector `gorm:"type:vector(1536)"`
}

type UserFile struct {
	MyModel
	UserID uint64 `gorm:"bigint;index;not null"`
	FileID uint64 `gorm:"bigint;index;not null"`
}

type Vector []float32

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
