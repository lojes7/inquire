package secure

import "golang.org/x/crypto/bcrypt"

// HashString 哈希加密函数
func HashString(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// VerifyPassword 验证是否匹配 若返回nil则无问题
func VerifyPassword(hashedPassword, inputPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(inputPassword))
}
