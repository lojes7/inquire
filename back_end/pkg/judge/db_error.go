package judge

import (
	"errors"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

// IsNotFound 判断是否是“记录不存在” 若为真则说明不存在
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsUniqueConflict 判断是否是“唯一约束冲突” 若为真则说明冲突
func IsUniqueConflict(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
