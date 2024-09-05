package utils

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Contains checks if a string is present in a slice of strings
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func IfElse(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

func ConvertValueFromString(value string, toDataType string) interface{} {
	switch toDataType {
	case "float64":
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	case "int":
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return value
}

func RemoveAccentsAndSpecialChars(input string) string {
	t := norm.NFD.String(input)
	var builder strings.Builder
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(builder.String(), "")
}

func GetString(m map[string]interface{}, key string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}

func GetStringSlice(m map[string]interface{}, key string) []string {
	if value, ok := m[key].([]interface{}); ok {
		result := make([]string, len(value))
		for i, v := range value {
			if s, ok := v.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return nil
}

func NormalizeAndFormat(s string) string {
	result := RemoveAccentsAndSpecialChars(s)
	result = strings.ToLower(result)
	result = strings.ReplaceAll(result, "/", "")
	result = strings.ReplaceAll(result, " ", "")
	return result
}
