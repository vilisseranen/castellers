package model

import (
	"testing"
)

func TestRoles(t *testing.T) {
	var emptyRoles []string
	validRoles := []string{"segon", "baix", "primera mà", "segona mà"}
	validRoles2 := []string{"acotxador"}
	invalidRoles := []string{"xxxxx", "segon", "baix"}
	invalidRoles2 := []string{"segon", "segon", "baix"}

	testEmpty := ValidateRoles(emptyRoles)
	if testEmpty != nil {
		t.Errorf("An empty role list should be valid.")
	}

	testValid := ValidateRoles(validRoles)
	if testValid != nil {
		t.Errorf("This list of roles should be valid: %v", validRoles)
	}

	testValid2 := ValidateRoles(validRoles2)
	if testValid2 != nil {
		t.Errorf("This list of roles should be valid: %v", validRoles2)
	}

	testInvalid := ValidateRoles(invalidRoles)
	if testInvalid == nil {
		t.Errorf("This list of roles should be invalid: %v", invalidRoles)
	}

	testInvalid2 := ValidateRoles(invalidRoles2)
	if testInvalid2 == nil {
		t.Errorf("This list of roles should be invalid: %v", invalidRoles2)
	}
}

func TestLanguages(t *testing.T) {
	emptyLanguage := ""
	validLanguage := "cat"
	invalidLanguage := "it"
	testEmpty := ValidateLanguage(emptyLanguage)
	if testEmpty == nil {
		t.Errorf("An empty Language should be invalid.")
	}

	testValid := ValidateLanguage(validLanguage)
	if testValid != nil {
		t.Errorf("This language should be valid: %v", validLanguage)
	}

	testInvalid := ValidateLanguage(invalidLanguage)
	if testInvalid == nil {
		t.Errorf("This language should be invalid: %v", invalidLanguage)
	}
}
