package utils

import (
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
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

func RemoveAccents(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
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
	result := RemoveAccents(s)
	result = strings.ToLower(result)
	result = strings.ReplaceAll(result, "/", "")
	result = strings.ReplaceAll(result, " ", "")
	return result
}
