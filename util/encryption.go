package util

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

func Encrypt(key string, plainText string) (string, error) {
	var cipherText []byte
	msg := []byte(plainText)

	aead, err := chacha20poly1305.NewX([]byte(key))
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(msg)+aead.Overhead())
	for i := 0; i < len(nonce); i++ {
		nonce[i] = byte(RandInt(0, 255))
		time.Sleep(time.Nanosecond)
	}
	cipherText = aead.Seal(nonce, nonce, msg, nil)

	return hex.EncodeToString(cipherText), nil
}

func Decrypt(key string, cipherText string) (string, error) {
	var plainText []byte

	cipherTextDecoded, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	aead, err := chacha20poly1305.NewX([]byte(key))
	if err != nil {
		return "", err
	}

	if len(cipherTextDecoded) < aead.NonceSize() {
		return "", err
	}

	nonce, ciphertext := cipherTextDecoded[:aead.NonceSize()], cipherTextDecoded[aead.NonceSize():]
	plainText, err = aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func ArgonHash(password string, salt string) string {
	return string(argon2.IDKey([]byte(password), []byte(salt), 1, 64*1024, 4, 32))
}

func Encode(text string) string {
	return hex.EncodeToString([]byte(text))
}

func Decode(encodedText string) (string, error) {
	decodedString, err := hex.DecodeString(encodedText)
	return string(decodedString), err
}
