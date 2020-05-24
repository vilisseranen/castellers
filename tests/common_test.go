package tests

import (
	"fmt"
	"testing"

	"github.com/vilisseranen/castellers/common"
)

func TestEncryption(t *testing.T) {
	originalText := "This is a text"
	encryptedText := common.Encrypt(originalText)
	decryptedText := common.Decrypt(encryptedText)
	if fmt.Sprintf("%s", encryptedText) == originalText {
		t.Error("Encrypted text is the same as plaintext.")
	}
	if decryptedText != originalText {
		t.Error("Decrypted text is different from original text.")
	}
}

func TestEncryption2(t *testing.T) {
	originalText := ""
	encryptedText := common.Encrypt(originalText)
	decryptedText := common.Decrypt(encryptedText)
	if fmt.Sprintf("%s", encryptedText) == originalText {
		t.Error("Encrypted text is the same as plaintext.")
	}
	if decryptedText != originalText {
		t.Error("Decrypted text is different from original text.")
	}
}

func TestHashing(t *testing.T) {
	password := "my super password"
	hashedPassword, _ := common.GenerateFromPassword(password)
	match := common.CompareHashAndPassword(hashedPassword, password)
	if fmt.Sprintf("%s", hashedPassword) == password {
		t.Error("Password was not encrypted.")
	}
	if match != nil {
		t.Error("Password was not decrypted properly.")
	}
}
