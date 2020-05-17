package tests

import (
	"testing"

	"github.com/vilisseranen/castellers/common"
)

func TestEncryption(t *testing.T) {
	originalText := "This is a text"
	encryptedText := common.Encrypt(originalText, common.GetConfigString("encryption_key"))
	decryptedText := common.Decrypt(encryptedText, common.GetConfigString("encryption_key"))
	if encryptedText == originalText {
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
	if hashedPassword == password {
		t.Error("Password was not encrypted.")
	}
	if match != nil {
		t.Error("Password was not decrypted properly.")
	}
}
