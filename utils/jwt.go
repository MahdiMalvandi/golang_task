package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

func base64UrlEncode(data []byte) string {
	enc := base64.URLEncoding.WithPadding(base64.NoPadding)
	return enc.EncodeToString(data)
}

func base64UrlDecode(encoded string) ([]byte, error) {
	enc := base64.URLEncoding.WithPadding(base64.NoPadding)
	return enc.DecodeString(encoded)
}


// Create JWT function
func CreateJwt(userID uint) (string, error) {
	// Create header and marshal
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJson, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	// Create payload and marshal
	payload := map[string]interface{}{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Encode header and payload
	encodedHeader := base64UrlEncode(headerJson)
	encodedPayload := base64UrlEncode(payloadJson)

	unsignedToken := encodedHeader + "." + encodedPayload


	// Hash token
	h := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET")))
	h.Write([]byte(unsignedToken))
	signature := base64UrlEncode(h.Sum(nil))

	return unsignedToken + "." + signature, nil
}

func VerifyJwt(token string) (userID uint, ok bool, err error) {
	// Split token and decode 
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, false, errors.New("invalid token")
	}

	unsignedToken := parts[0] + "." + parts[1]
	signature := parts[2]

	payloadJson, err := base64UrlDecode(parts[1])
	if err != nil {
		return 0, false, err
	}

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payloadJson, &payloadMap); err != nil {
		return 0, false, err
	}

	// Check expire time
	var tokenExpireTime = int64(payloadMap["exp"].(float64))
	if tokenExpireTime <= time.Now().Unix() {
		return 0, false, errors.New("token had expired")
	}

	// Check Token
	h := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET")))
	h.Write([]byte(unsignedToken))
	expectedSignature := base64UrlEncode(h.Sum(nil))

	check := hmac.Equal([]byte(signature), []byte(expectedSignature))
	if check {
		userID = uint(payloadMap["sub"].(float64))
	} else {
		userID = 0
	}
	return userID, check, nil
}
