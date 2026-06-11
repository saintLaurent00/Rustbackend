package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

func Encrypt(text string, keyString string) (string, error) {
	key := []byte(keyString)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plaintext := []byte(text)
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	return hex.EncodeToString(append(nonce, ciphertext...)), nil
}

func Decrypt(encryptedText string, keyString string) (string, error) {
	key := []byte(keyString)
	data, err := hex.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	if len(data) < 12 {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce := data[:12]
	ciphertext := data[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
