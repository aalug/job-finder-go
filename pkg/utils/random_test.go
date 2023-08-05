package utils

import (
	"github.com/aalug/go-gin-job-search/pkg/validation"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestRandomInt(t *testing.T) {
	// Set the seed
	rand.Seed(time.Now().UnixNano())

	// Test with min and max as 0
	min := int32(0)
	max := int32(0)
	randomValue := RandomInt(min, max)
	require.Equal(t, int32(0), randomValue, "Random value should be 0 when min and max are both 0")

	// Test with a positive range
	min = int32(10)
	max = int32(20)
	randomValue = RandomInt(min, max)
	require.True(t, min <= randomValue && randomValue <= max)

	// Test with a negative range
	min = int32(-100)
	max = int32(-50)
	randomValue = RandomInt(min, max)
	require.True(t, min <= randomValue && randomValue <= max)

	// Test with a large range
	min = int32(0)
	max = int32(1000000)
	randomValue = RandomInt(min, max)
	require.True(t, min <= randomValue && randomValue <= max)
}

func TestRandomString(t *testing.T) {
	// Set the seed based on the current time to get different values for each test run
	rand.Seed(time.Now().UnixNano())

	// Test with n = 0
	n := 0
	randomString := RandomString(n)
	require.Equal(t, "", randomString, "RandomString should return an empty string when n is 0")

	// Test with n = 10
	n = 10
	randomString = RandomString(n)
	require.Equal(t, n, len(randomString))

	// Test with a large value of n
	n = 1000
	randomString = RandomString(n)
	require.Equal(t, n, len(randomString))

	// Test with the alphabet characters
	randomString = RandomString(len(alphabet))
	require.True(t, isStringInAlphabet(randomString), "RandomString should only contain characters from the alphabet")

	// Test with a very large value of n
	n = 1000000
	randomString = RandomString(n)
	require.Equal(t, n, len(randomString))
	require.True(t, isStringInAlphabet(randomString), "RandomString should only contain characters from the alphabet")
}

func TestRandomEmail(t *testing.T) {
	// Test with multiple random emails
	for i := 0; i < 10; i++ {
		randomEmail := RandomEmail()

		// Check if the email has the correct format
		err := validation.ValidateEmail(randomEmail)
		require.NoError(t, err)
	}
}

func isStringInAlphabet(str string) bool {
	for _, c := range str {
		if !strings.ContainsRune(alphabet, c) {
			return false
		}
	}
	return true
}
