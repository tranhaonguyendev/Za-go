package util

import "unicode/utf16"

func UTF16Len(text string) int {
	return len(utf16.Encode([]rune(text)))
}

func UTF16RuneLen(r rune) int {
	if r <= 0xFFFF {
		return 1
	}
	return 2
}
