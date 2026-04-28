package model

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/lojes7/inquire/pkg/utils"
	"gorm.io/gorm"
)

type FileVector struct {
	MyModel
	FileID uint64 `gorm:"type:bigint;not null;index"`
	Vector Vector `gorm:"type:vector(1024);not null"`
	Number int    `gorm:"type:integer;not null"`
}

// Vector 用于映射 pgvector 向量字段。
type Vector []float32

func (fv *FileVector) BeforeCreate(db *gorm.DB) error {
	if fv.ID == 0 {
		fv.ID = utils.NewUniqueID()
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
