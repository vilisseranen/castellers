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

func TestNumberOrEmpty(t *testing.T) {
	empty := ""
	integer := "10"
	decimal := "12.10"
	withComma := "11,10"
	withTooManyDecimals := "1.10000"

	testEmpty := ValidNumberOrEmpty(empty)
	if testEmpty != nil {
		t.Errorf("An empty field should be valid.")
	}

	testInteger := ValidNumberOrEmpty(integer)
	if testInteger != nil {
		t.Errorf("An integer should be valid.")
	}

	testDecimal := ValidNumberOrEmpty(decimal)
	if testDecimal != nil {
		t.Errorf("A decimal should be valid.")
	}

	testWithComma := ValidNumberOrEmpty(withComma)
	if testWithComma == nil {
		t.Errorf("A number with comma should be invalid.")
	}

	testWithTooManyDecimals := ValidNumberOrEmpty(withTooManyDecimals)
	if testWithTooManyDecimals == nil {
		t.Errorf("A number with more than 2 decimals should be invalid.")
	}
}
