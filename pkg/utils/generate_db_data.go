package utils

import "fmt"

var Industries = []string{
	"IT",
	"Game Development",
	"Financial Technology",
	"E-commerce",
}

var Locations = []string{
	"London",
	"New York",
	"Paris",
	"Berlin",
	"Madrid",
	"Milan",
	"Rome",
	"Amsterdam",
	"Barcelona",
	"Vienna",
	"Hamburg",
	"Dublin",
	"Sydney",
	"Singapore",
	"Hong Kong",
	"Tokyo",
	"Seoul",
	"Warsaw",
}

// qualifiers
var qualifiers = []string{"Junior", "Senior", "Lead"}

// languages for Developers and Engineers
var languages = []string{"Java", "Python", "JavaScript", "Go", "C#", "C++", "Ruby", "C"}

// GenerateDeveloperJobs Function to combine different parts for Developers
func GenerateDeveloperJobs() []string {
	var combined []string
	for _, qualifier := range qualifiers {
		for _, language := range languages {
			combined = append(combined, fmt.Sprintf("%s %s Developer", qualifier, language))
		}
	}
	return combined
}

// GenerateEngineerJobs Function to combine different parts for Engineers
func GenerateEngineerJobs() []string {
	var combined []string
	for _, qualifier := range qualifiers {
		for _, language := range languages {
			combined = append(combined, fmt.Sprintf("%s %s Engineer", qualifier, language))
		}
	}
	return combined
}
