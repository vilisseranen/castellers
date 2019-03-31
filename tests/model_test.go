package tests

import (
	"testing"

	"github.com/vilisseranen/castellers/model"
)

func TestRoles(t *testing.T) {
	var emptyRoles []string
	validRoles := []string{"segon", "baix", "primera mà", "segona mà"}
	validRoles2 := []string{"acotxador"}
	invalidRoles := []string{"xxxxx", "segon", "baix"}
	invalidRoles2 := []string{"segon", "segon", "baix"}

	testEmpty := model.ValidateRoles(emptyRoles)
	if testEmpty != nil {
		t.Errorf("An empty role list should be valid.")
	}

	testValid := model.ValidateRoles(validRoles)
	if testValid != nil {
		t.Errorf("This list of roles should be valid: %v", validRoles)
	}

	testValid2 := model.ValidateRoles(validRoles2)
	if testValid2 != nil {
		t.Errorf("This list of roles should be valid: %v", validRoles2)
	}

	testInvalid := model.ValidateRoles(invalidRoles)
	if testInvalid == nil {
		t.Errorf("This list of roles should be invalid: %v", invalidRoles)
	}

	testInvalid2 := model.ValidateRoles(invalidRoles2)
	if testInvalid2 == nil {
		t.Errorf("This list of roles should be invalid: %v", invalidRoles2)
	}
}

func TestLanguages(t *testing.T) {
	emptyLanguage := ""
	validLanguage := "cat"
	invalidLanguage := "it"
	testEmpty := model.ValidateLanguage(emptyLanguage)
	if testEmpty == nil {
		t.Errorf("An empty Language should be invalid.")
	}

	testValid := model.ValidateLanguage(validLanguage)
	if testValid != nil {
		t.Errorf("This language should be valid: %v", validLanguage)
	}

	testInvalid := model.ValidateLanguage(invalidLanguage)
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

	testEmpty := model.ValidNumberOrEmpty(empty)
	if testEmpty != nil {
		t.Errorf("An empty field should be valid.")
	}

	testInteger := model.ValidNumberOrEmpty(integer)
	if testInteger != nil {
		t.Errorf("An integer should be valid.")
	}

	testDecimal := model.ValidNumberOrEmpty(decimal)
	if testDecimal != nil {
		t.Errorf("A decimal should be valid.")
	}

	testWithComma := model.ValidNumberOrEmpty(withComma)
	if testWithComma == nil {
		t.Errorf("A number with comma should be invalid.")
	}

	testWithTooManyDecimals := model.ValidNumberOrEmpty(withTooManyDecimals)
	if testWithTooManyDecimals == nil {
		t.Errorf("A number with more than 2 decimals should be invalid.")
	}
}
