package service

import (
	"vvechat/internal/model"
	"vvechat/pkg/infra"
)

func RefreshToken(id uint64) (*model.TokenResp, error) {
	return NewTokenResp(id)
}

func ReviseUid(id uint64, newUid string) error {
	return infra.GetDB().
		Model(&model.User{}).
		Where("id = ?", id).
		Update("uid", newUid).
		Error
}
