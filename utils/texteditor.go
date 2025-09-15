package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// Разделяет по любым пробелам
func SplitTextToThreeVars(text string) (string, string, string, error) {
	parts := strings.Fields(text)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("Получено больше или меньше 3 слов, а получено : %d", len(parts))
	}
	return parts[0], parts[1], parts[2], nil
}

func IsValidSingleWord(text string) (bool, string) {
	text = strings.TrimSpace(text)

	if text == "" {
		return false, "Username не может быть пустым"
	}

	if strings.ContainsAny(text, " \t\n") {
		return false, "Username не должен содержать пробелы"
	}

	if len(text) > 50 {
		return false, "Username не должен превышать 50 символов"
	}

	// Проверяем что содержит только английские буквы и цифры
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, text)
	if !matched {
		return false, "Допускаются только английские буквы, цифры и нижнее подчеркивание"
	}

	return true, ""
}
