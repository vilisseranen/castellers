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

func createHash(key, salt string) []byte {
	iterations := GetConfigInt("encryption.iterations")
	return pbkdf2.Key([]byte(key), []byte(salt), iterations, 32, sha256.New)
}

func Encrypt(data string, passphrase string) string {
	salt := GetConfigString("encryption.key_salt")
	block, _ := aes.NewCipher([]byte(createHash(passphrase, salt)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	return fmt.Sprintf("%s", ciphertext)
}

func Decrypt(data string, passphrase string) string {
	salt := GetConfigString("encryption.key_salt")
	key := []byte(createHash(passphrase, salt))
	block, err := aes.NewCipher(key)
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

func GenerateFromPassword(password string) (string, error) {
	pepper := GetConfigString("encryption.password_pepper")
	cost := GetConfigInt("encryption.password_hashing_cost")
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password+pepper), cost)
	return fmt.Sprintf("%s", hashedPasswordBytes), err
}

func CompareHashAndPassword(hashedPassword, password string) error {
	pepper := GetConfigString("encryption.password_pepper")
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password+pepper))
}
