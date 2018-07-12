package model

import (
	"testing"
)

func TestRoles(t *testing.T) {
	emptyRoles := ""
	validRoles := "segond,baix"
	invalidRoles := "segond,toto"

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
}
