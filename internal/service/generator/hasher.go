package generator

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GetHashWithRandomSalt(input string) (string, string, error) {

	salt := make([]byte, 16) // 16 bytes = 32 hex chars

	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}

	hashHex, saltHex := GetHashWithSalt(input, salt)

	return hashHex, saltHex, nil
}

func GetHashWithSalt(input string, salt []byte) (string, string) {
	mac := hmac.New(sha256.New, []byte(salt))
	mac.Write([]byte(input))
	return hex.EncodeToString(mac.Sum(nil)), hex.EncodeToString([]byte(salt))
}
