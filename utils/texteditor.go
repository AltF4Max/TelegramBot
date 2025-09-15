package utils

import (
	"fmt"
	"strings"
)

// Разделяет по любым пробелам
func SplitTextToThreeVars(text string) (string, string, string, error) {
	parts := strings.Fields(text)
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("нужно ввести 3 слова, а получено %d", len(parts))
	}
	return parts[0], parts[1], parts[2], nil
}
