package common

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
)

const UUID_SIZE = 40
const CODE_SIZE = 16

func GenerateUUID() string {
	data := make([]byte, 10)
	_, err := rand.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))[:UUID_SIZE]
}

func GenerateCode() string {
	data := make([]byte, 10)
	_, err := rand.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))[:CODE_SIZE]
}

const ANSWER_YES = "yes"
const ANSWER_NO = "no"
const ANSWER_MAYBE = "maybe"
