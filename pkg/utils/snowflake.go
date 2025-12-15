package utils

import (
	"errors"

	"github.com/sony/sonyflake"
)

func NextUniqueID() (uint64, error) {
	sf := sonyflake.NewSonyflake(sonyflake.Settings{})
	if sf == nil {
		return 0, errors.New("创建sonyflake实例失败")
	}
	return sf.NextID()
}
