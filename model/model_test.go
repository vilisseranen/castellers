package model

import (
	"testing"
)

func TestRoles(t *testing.T) {
	var emptyRoles []string
	validRoles := []string{"segon", "baix", "primera mà", "segona mà"}
	invalidRoles := []string{"segon", "toto", "baix"}
	invalidRoles2 := []string{"segon", "segon", "baix"}

	testEmpty := ValidateRoles(emptyRoles)
	if testEmpty != nil {
		t.Errorf("An empty role list should be valid.")
	}

	testValid := ValidateRoles(validRoles)
	if testValid != nil {
		t.Errorf("This list of roles should be valid: %v", validRoles)
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
