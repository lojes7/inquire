package secure

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

var (
	jwtKey      []byte
	expiresTime time.Duration
	refreshTime time.Duration
)

func InitJWT() error {
	key := viper.GetString("jwt.key")
	if key == "" {
		return errors.New("jwtKey 是空的")
	}

	jwtKey = []byte(key)
	expiresTime = viper.GetDuration("jwt.expires_time")
	refreshTime = viper.GetDuration("jwt.refresh_time")
	return nil
}

func GetExpiresTime() time.Duration {
	return expiresTime
}

func GetRefreshTime() time.Duration {
	return refreshTime
}

type IDClaims struct {
	ID   uint64 `json:"id"`
	Type string `json:"type"`
	jwt.RegisteredClaims
}

func generateToken(id uint64, t time.Duration, tokenType string) (string, error) {
	if string(jwtKey) == "" {
		return "", errors.New("jwtKey在生成Token时发现是空的")
	}

	claims := IDClaims{
		ID:   id,
		Type: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并生成字符串
	return token.SignedString(jwtKey)
}

func NewToken(id uint64) (string, error) {
	return generateToken(id, expiresTime, "access")
}

func NewRefreshToken(id uint64) (string, error) {
	return generateToken(id, refreshTime, "refresh")
}

func ParseToken(tokenString string) (*IDClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &IDClaims{},
		func(token *jwt.Token) (any, error) {
			return jwtKey, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*IDClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token不合法")
	}

	return claims, nil
}
