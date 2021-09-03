package model

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
)

var ValidRoleList = []string{
	"baix",
	"contrafort",
	"primera mà",
	"segona mà",
	"lateral",
	"vent",
	"agulla",
	"crossa",
	"segon",
	"terç",
	"quart",
	"dos",
	"acotxador",
	"enxaneta",
	"pinya",
}

var ValidLanguageList = []string{
	"fr",
	"en",
	"cat",
}

var ValidEventTypes = []string{
	"presentation",
	"practice",
	"social",
}

var ValidMemberTypes = []string{
	MEMBERSTYPEADMIN,
	MEMBERSTYPEREGULAR,
	MEMBERSTYPEGUEST,
}

func ValidateRoles(roles []string) error {
	sort.Strings(roles)
	sort.Strings(ValidRoleList)
	validRoles := ValidRoleList
	index := -1
	for _, roleToTest := range roles {
		validRoles = validRoles[index+1:]
		index = sort.SearchStrings(validRoles, roleToTest)
		if index == len(validRoles) || roleToTest != validRoles[index] {
			return errors.New("Invalid roles")
		}
	}
	return nil
}

func ValidNumberOrEmpty(field string) error {
	// This regex matches:
	// - empty strings (not required fields)
	// - numbers without digits, ex: 180
	// - numbers with 2 decimals, ex: 180.20
	re := regexp.MustCompile(`^(\d+(\.\d{1,2})?)?$`)
	if !re.MatchString(field) {
		return errors.New(fmt.Sprintf("%v is not a valid number", field))
	}
	return nil
}

func ValidateLanguage(language string) error {
	sort.Strings(ValidLanguageList)
	index := sort.SearchStrings(ValidLanguageList, language)
	if index == len(ValidLanguageList) || language != ValidLanguageList[index] {
		return errors.New("Invalid language")
	}
	return nil
}

func ValidateType(memberType string) error {
	sort.Strings(ValidMemberTypes)
	index := sort.SearchStrings(ValidMemberTypes, memberType)
	if index == len(ValidMemberTypes) || memberType != ValidMemberTypes[index] {
		return errors.New("Invalid type")
	}
	return nil
}
