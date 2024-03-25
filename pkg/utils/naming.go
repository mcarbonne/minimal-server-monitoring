package utils

import "regexp"

func IsNameValid(word string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_-]*$`).MatchString(word)
}
