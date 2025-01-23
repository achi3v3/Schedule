package functions

import (
	"regexp"
	"strconv"
	"strings"
)

func renameSheetGroup(s string) string {
	if checkString(string(s[0])) {
		s = string(s[1:]) + string(s[0])
	}
	return strings.ToLower(replaceCyrillicWithLatin(removeSpaces(removeDots(s))))
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
func checkString(s string) bool {
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}
func isValidFormat(s string) bool {
	re := regexp.MustCompile(`^[\wа-яА-ЯёЁ-]+(?:-[\wа-яА-ЯёЁ-]+)+$`)
	return re.MatchString(s)
}
func replaceCyrillicWithLatin(input string) string {
	translitMap := map[rune]string{
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "E", 'Ж': "ZH", 'З': "Z", 'И': "I",
		'Й': "I", 'К': "K", 'Л': "L", 'М': "M", 'Н': "N", 'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T", 'У': "U",
		'Ф': "F", 'Х': "H", 'Ц': "TS", 'Ч': "CH", 'Ш': "SH", 'Щ': "SHH", 'Ы': "Y", 'Э': "E", 'Ю': "YU", 'Я': "YA",
		// Маленькие буквы кириллицы
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh", 'з': "z", 'и': "i",
		'й': "i", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
		'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shh", 'ы': "y", 'э': "e", 'ю': "yu", 'я': "ya",
	}
	var result strings.Builder
	for _, ch := range input {
		if ch == '-' || ch == '(' || ch == ')' {
			continue
		}
		if replacement, ok := translitMap[ch]; ok {
			result.WriteString(replacement)
		} else {
			result.WriteRune(ch)
		}
	}
	return result.String()
}
