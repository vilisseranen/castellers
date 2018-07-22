package model

import (
	"errors"
	"strings"
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
	"segond",
	"terç",
	"quart",
	"dosos",
	"acotxador",
	"enxaneta",
	"pinya",
}

func ValidateRole(roleString string) error {
	if roleString == "" {
		return nil
	}
	roleList := strings.Split(roleString, ",")
	for _, role := range roleList {
		if strings.Contains(","+strings.Join(ValidRoleList, ",")+",", ","+role+",") == false {
			return errors.New("Invalid roles")
		}
		if strings.Count(strings.Join(roleList, ""), role) > 1 {
			return errors.New("Invalid roles")
		}
	}
	return nil
}
