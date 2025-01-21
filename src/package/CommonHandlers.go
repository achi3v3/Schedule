package funcExcel

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func FunctionCommonHandlers() {
	fmt.Println("func: Common Handlers")
}

// ================================================FIND FUNCS==============================================================================
func findAndAdd(slice [][]string, value string, newValue string) ([][]string, bool) {
	for i, subSlice := range slice {
		if contains(subSlice, value) {
			slice[i] = append(slice[i], newValue)
			return slice, true
		}
	}
	return slice, false
}
func findIndex(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}
func findListContainingString(listOfLists [][]string, target string) []string {
	for _, innerList := range listOfLists {
		for _, item := range innerList {
			if item == target {
				return innerList
			}
		}
	}
	return nil
}
func checkString(s string) bool {
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}
func isValidFormat(s string) bool {
	// Регулярное выражение для строки с хотя бы одним дефисом
	re := regexp.MustCompile(`^[\wа-яА-ЯёЁ-]+(?:-[\wа-яА-ЯёЁ-]+)+$`)
	return re.MatchString(s)
}

// ================================================CONTAIN FUNCS==============================================================================
func contains(slice []string, value string) bool {
	normalizedValue := strings.TrimSpace(value)
	for _, v := range slice {
		normalizedSliceValue := strings.TrimSpace(v)
		if normalizedValue == normalizedSliceValue {
			return true
		}
	}
	return false
}
func containsInNested(slice [][]string, value string) bool {
	for _, subSlice := range slice {
		if contains(subSlice, value) {
			return true
		}
	}
	return false
}

// ================================================REMOVE FUNCS==============================================================================
func removeExtraSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
func removeSpaces(input string) string {
	return strings.ReplaceAll(input, " ", "")
}
func removeDots(input string) string {
	return strings.ReplaceAll(input, ".", "")
}
func removeDuplicatesLists(input []string) []string {
	// Создаем map для отслеживания уникальных элементов
	seen := make(map[string]bool)
	var result []string

	// Проходим по всем элементам входного списка
	for _, item := range input {
		// Если элемент еще не добавлен в map, добавляем его
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	// Возвращаем новый срез с уникальными элементами
	return result
}

// ================================================GET FUNCS==============================================================================
func get_days_for_couple() []string {
	return []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}
}
