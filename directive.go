package template

import (
	"errors"
	"strings"
)

// Parse and extract directive (expect what string begins with leftDelim).
func extractDirective(s *string) (directive string, err error) {
	i := strings.Index(*s, rightDelim)
	if i == -1 {
		return "", errors.New("directive with no end")
	}
	directive = (*s)[len(leftDelim):i]
	*s = (*s)[i+len(rightDelim):]
	return
}

// Same as extractDirective but keep 's' as-is.
func getDirective(s *string) (directive string, err error) {
	i := strings.Index(*s, rightDelim)
	if i == -1 {
		return "", errors.New("directive with no end")
	}
	directive = (*s)[len(leftDelim):i]
	return
}
