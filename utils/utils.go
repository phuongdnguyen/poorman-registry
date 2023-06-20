package utils

import (
	"errors"
	"net/http"
	"strings"
)

func Redact(in http.Header) http.Header {
	h := in.Clone()
	if h.Get("Authorization") != "" {
		h.Set("Authorization", "REDACTED")
	}
	return h
}

// SelectNotEmpty return 1 element if there is 1 empty element
// SelectNotEmpty return 1 element if 2 element are equal
// SelectNotEmpty return error otherwise
func SelectNotEmpty(a, b string) (string, error) {
	if a == "" {
		return b, nil
	}
	if b == "" {
		return a, nil
	}
	if a != b {
		return "", errors.New("argument not equal")
	}
	if a == b {
		return a, nil
	}
	return "", errors.New("both argument must not be empty")
}

func SplitAndGetLast(delimiter, input string) string {
	return input[strings.LastIndex(input, delimiter)+1:]
}

func MakeImageName(name, tag string) string {
	return name + ":" + tag
}

func MakeRepoName(unEncodedRegistry string) string {
	return strings.ReplaceAll(unEncodedRegistry, ".", "-")
}
