package model

import (
	"errors"
	"sort"
)

var ValidLanguageList = []string{
	"fr",
	"en",
	"cat",
}

func ValidateLanguage(language string) error {
	sort.Strings(ValidLanguageList)
	index := sort.SearchStrings(ValidLanguageList, language)
	if index == len(ValidLanguageList) || language != ValidLanguageList[index] {
		return errors.New("Invalid language")
	}
	return nil
}
