package common

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

const uuidSize = 40
const codeSize = 16

const AnswerYes = "yes"
const AnswerNo = "no"
const AnswerMaybe = "maybe"

func GenerateUUID() string {
	data := make([]byte, 10)
	_, err := rand.Read(data)
	if err != nil {
		Fatal(err.Error())
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))[:uuidSize]
}

func GenerateCode() string {
	data := make([]byte, 10)
	_, err := rand.Read(data)
	if err != nil {
		Fatal(err.Error())
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))[:codeSize]
}
