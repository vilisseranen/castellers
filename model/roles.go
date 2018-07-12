package model

import (
	"errors"
	"strings"
)

var validRoleList = []string{
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
	valid := true
	roleString = strings.Replace(roleString, ", ", ",", -1)
	roleList := strings.Split(roleString, ",")
	for _, role := range roleList {
		valid = false
		for _, valideRole := range validRoleList {
			if valideRole == role {
				valid = true
			}
		}
	}
	if valid == false {
		return errors.New("Invalid roles")
	} else {
		return nil
	}
}
