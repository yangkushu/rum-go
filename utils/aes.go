package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// AesEncrypt 加密数据
func AesEncrypt(text string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nil, nonce, []byte(text), nil)
	authTag := ciphertext[len(ciphertext)-aesGCM.Overhead():]
	ciphertextWithoutTag := ciphertext[:len(ciphertext)-aesGCM.Overhead()]

	enc := base64.StdEncoding.EncodeToString(ciphertextWithoutTag)
	iv := base64.StdEncoding.EncodeToString(nonce)
	tag := base64.StdEncoding.EncodeToString(authTag)

	return enc + "~" + iv + "~" + tag, nil
}

// AesDecrypt 解密数据
func AesDecrypt(encryptedData string, key []byte) (string, error) {
	parts := strings.Split(encryptedData, "~")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid ciphertext format")
	}

	enc, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", err
	}
	iv, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	authTag, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCMWithTagSize(block, len(authTag))
	if err != nil {
		return "", err
	}

	aesgcm.Seal(nil, iv, nil, nil) // Only to set the tag size
	ciphertext := append(enc, authTag...)

	plaintext, err := aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
