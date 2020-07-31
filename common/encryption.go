package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

var encryption_key []byte

func getKey() []byte {
	if encryption_key == nil {
		key := GetConfigString("encryption.key")
		salt := GetConfigString("encryption.key_salt")
		iterations := GetConfigInt("encryption.iterations")
		encryption_key = pbkdf2.Key([]byte(key), []byte(salt), iterations, 32, sha256.New)
	}
	return encryption_key
}

func Encrypt(data string) []byte {
	block, _ := aes.NewCipher(getKey())
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	return ciphertext
}

func Decrypt(data []byte) string {
	block, err := aes.NewCipher(getKey())
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := []byte(data)[:nonceSize], []byte(data)[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, []byte(ciphertext), nil)
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("%s", plaintext)
}

func GenerateFromPassword(password string) ([]byte, error) {
	pepper := GetConfigString("encryption.password_pepper")
	cost := GetConfigInt("encryption.password_hashing_cost")
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password+pepper), cost)
	return hashedPasswordBytes, err
}

func CompareHashAndPassword(hashedPassword []byte, password string) error {
	pepper := GetConfigString("encryption.password_pepper")
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password+pepper))
}
