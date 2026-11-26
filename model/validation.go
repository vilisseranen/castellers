package model

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"
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

// EVENTSTYPEMUSIC is a rehearsal of the musicians of the colla. Its emails
// only go to the members who are musicians.
const EVENTSTYPEMUSIC = "music"

var ValidEventTypes = []string{
	"presentation",
	"practice",
	"social",
	EVENTSTYPEMUSIC,
}

var ValidMemberTypes = []string{
	MEMBERSTYPEADMIN,
	MEMBERSTYPEREGULAR,
	MEMBERSTYPEGUEST,
	MEMBERSTYPECANALLA,
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

// Sizes of the shirts of the colla: even numbers only.
const MINSHIRTSIZE = 28
const MAXSHIRTSIZE = 44

// SHIRTSINCEFORMAT is how the day a member got their shirt is written.
const SHIRTSINCEFORMAT = "2006-01-02"

// ValidateShirt checks the shirt of a member: the size is an even number from
// MINSHIRTSIZE to MAXSHIRTSIZE, or 0 when it is not known; since is a day that
// is not in the future, or empty when it is not known.
func ValidateShirt(size int, since string, now time.Time) error {
	if size != 0 && (size < MINSHIRTSIZE || size > MAXSHIRTSIZE || size%2 != 0) {
		return errors.New("Invalid shirt size")
	}
	if since != "" {
		day, err := time.Parse(SHIRTSINCEFORMAT, since)
		// One day of margin: it may already be tomorrow where the admin is.
		if err != nil || day.After(now.Add(24*time.Hour)) {
			return errors.New("Invalid shirt date")
		}
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

// Reasons why an event is cancelled.
const EVENTCANCELREASONWEATHER = "weather"
const EVENTCANCELREASONTECHNICAL = "technical"
const EVENTCANCELREASONOTHER = "other"

var ValidCancelReasons = []string{
	EVENTCANCELREASONWEATHER,
	EVENTCANCELREASONTECHNICAL,
	EVENTCANCELREASONOTHER,
}
