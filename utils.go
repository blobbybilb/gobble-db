package gobble

import "strings"

func isValidCollectionName(name string) bool {
	if len(name) == 0 {
		return false
	}

	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-_"
	for _, c := range name {
		if !strings.ContainsRune(validChars, c) {
			return false
		}
	}

	if strings.ContainsRune("-_", rune(name[0])) || strings.ContainsRune(".-_", rune(name[len(name)-1])) {
		return false
	}

	return true
}

func isValidIndexName(name string) bool {
	if len(name) == 0 {
		return false
	}

	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-_"
	for _, c := range name {
		if !strings.ContainsRune(validChars, c) {
			return false
		}
	}

	if strings.ContainsRune("-_", rune(name[0])) || strings.ContainsRune(".-_", rune(name[len(name)-1])) {
		return false
	}

	return true
}
