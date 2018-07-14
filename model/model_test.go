package model

import (
	"testing"
)

func TestRoles(t *testing.T) {
	emptyRoles := ""
	validRoles := "segond,baix,primera mà,segona mà"
	invalidRoles := "segond,toto,baix"
	invalidRoles2 := "segond,,baix"
	invalidRoles3 := "segond,segond,baix"

	testEmpty := ValidateRole(emptyRoles)
	if testEmpty != nil {
		t.Errorf("An empty role list should be valid.")
	}

	testValid := ValidateRole(validRoles)
	if testValid != nil {
		t.Errorf("This list of roles should be valid: %v", validRoles)
	}

	testInvalid := ValidateRole(invalidRoles)
	if testInvalid == nil {
		t.Errorf("This list of roles should be invalid: %v", invalidRoles)
	}

	testInvalid2 := ValidateRole(invalidRoles2)
	if testInvalid2 == nil {
		t.Errorf("This list of roles should be invalid: %v", invalidRoles2)
	}

	testInvalid3 := ValidateRole(invalidRoles3)
	if testInvalid3 == nil {
		t.Errorf("This list of roles should be invalid: %v", invalidRoles3)
	}
}
