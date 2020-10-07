package util

import "golang.org/x/crypto/ssh/terminal"

func IntMin(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func GetConsoleWidth() int {
	width, _, err := terminal.GetSize(0)
	if err != nil {
		width = 80
	}

	return width
}
