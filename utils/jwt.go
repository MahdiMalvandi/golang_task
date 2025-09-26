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

func CreateJwt(userId uint) (string, error) {
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJson, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	payload := map[string]interface{}{
		"sub": userId,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	encodedHeader := base64UrlEncode(headerJson)
	encodedPayload := base64UrlEncode(payloadJson)

	unsignedToken := encodedHeader + "." + encodedPayload

	h := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET")))
	h.Write([]byte(unsignedToken))
	signature := base64UrlEncode(h.Sum(nil))

	return unsignedToken + "." + signature, nil
}

func VerifyJwt(token string) (userID uint, ok bool, err error) {
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

	var tokenExpireTime = int64(payloadMap["exp"].(float64))
	if tokenExpireTime <= time.Now().Unix() {
		return 0, false, errors.New("token had expired")
	}
	h := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET")))
	h.Write([]byte(unsignedToken))
	expectedSignature := base64UrlEncode(h.Sum(nil))


	check := hmac.Equal([]byte(signature), []byte(expectedSignature))
	if check{
		userID = uint(payloadMap["sub"].(float64))
	} else {
		userID = 0
	}
	return userID, check, nil
}
