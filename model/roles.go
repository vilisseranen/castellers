package model

import (
	"errors"
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
	"dosos",
	"acotxador",
	"enxaneta",
	"pinya",
}

func ValidateRoles(roles []string) error {
	sort.Strings(roles)
	sort.Strings(ValidRoleList)
	validRoles := ValidRoleList[:]
	index := -1
	for _, roleToTest := range roles {
		validRoles = validRoles[index+1 : len(validRoles)]
		index = sort.SearchStrings(validRoles, roleToTest)
		if index == len(validRoles) || roleToTest != validRoles[index] {
			return errors.New("Invalid roles")
		}
	}
	return nil
}
