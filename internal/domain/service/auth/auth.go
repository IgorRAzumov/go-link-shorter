package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	secretKey string
}

func NewAuthService(secretKey string) *Service {
	return &Service{secretKey: secretKey}
}

func (service *Service) GenerateUserID() string {
	return uuid.NewString()
}

func (service *Service) SignUserID(userID string) string {
	signature := service.hmacHex(userID)
	return base64.URLEncoding.EncodeToString([]byte(userID)) + "." + signature
}

func (service *Service) ValidateSignedUserID(value string) (string, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid cookie format")
	}

	userIDBytes, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", err
	}
	userID := string(userIDBytes)

	expected := service.hmacHex(userID)
	if !hmac.Equal([]byte(parts[1]), []byte(expected)) {
		return "", fmt.Errorf("invalid cookie signature")
	}

	return userID, nil
}

func (service *Service) hmacHex(message string) string {
	mac := hmac.New(sha256.New, []byte(service.secretKey))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
