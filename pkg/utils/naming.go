package utils

import "regexp"

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]*$`)

func IsNameValid(word string) bool {
	return validNameRegex.MatchString(word)
}
