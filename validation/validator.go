package validation

import (
	"fmt"
	"net/mail"
)

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

	_, err := mail.ParseAddress(value)
	if err != nil {
		return fmt.Errorf("email is invalid")
	}

	return nil
}
