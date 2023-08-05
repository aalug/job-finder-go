package validation

import (
	"fmt"
	"regexp"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

// ValidateStringLength check if the string length is between minLength and maxLength
func ValidateStringLength(value string, minLength, maxLength int) error {
	n := len(value)
	if n < minLength || n > maxLength {
		return fmt.Errorf("string length is invalid, must be between %d and %d", minLength, maxLength)
	}

	return nil
}

// ValidateEmail check if the email is valid.
// It must be between 5 and 250 characters long
// and contain a valid email address.
func ValidateEmail(value string) error {
	if err := ValidateStringLength(value, 5, 250); err != nil {
		return err
	}

	if !emailRegex.MatchString(value) {
		return fmt.Errorf("email is invalid")
	}

	return nil
}
