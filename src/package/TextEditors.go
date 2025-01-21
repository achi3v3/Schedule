package funcExcel

import (
	"fmt"
	"strings"
)

func FunctionTextEditors() {
	fmt.Println("func Text Editors")
}

func renameSheetGroup(s string) string {
	if checkString(string(s[0])) {
		s = string(s[1:]) + string(s[0])
	}
	return strings.ToLower(replaceCyrillicWithLatin(removeSpaces(removeDots(s))))
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

func replaceLatinWithCyrillic(input string) string {
	translitMap := map[string]string{
		"A": "А", "B": "Б", "V": "В", "G": "Г", "D": "Д", "E": "Е", "ZH": "Ж", "Z": "З", "I": "И",
		"K": "К", "L": "Л", "M": "М", "N": "Н", "O": "О", "P": "П", "R": "Р", "S": "С", "T": "Т", "U": "У",
		"F": "Ф", "H": "Х", "TS": "Ц", "CH": "Ч", "SH": "Ш", "SHH": "Щ", "Y": "Ы", "YU": "Ю", "YA": "Я",
		// Маленькие буквы латиницы
		"a": "а", "b": "б", "v": "в", "g": "г", "d": "д", "e": "е", "zh": "ж", "z": "з", "i": "и",
		"k": "к", "l": "л", "m": "м", "n": "н", "o": "о", "p": "п", "r": "р", "s": "с", "t": "т", "u": "у",
		"f": "ф", "h": "х", "ts": "ц", "ch": "ч", "sh": "ш", "shh": "щ", "y": "ы", "yu": "ю", "ya": "я",
	}

	var result strings.Builder
	i := 0
	runes := []rune(input)

	for i < len(runes) {
		if i+1 < len(runes) {
			// Проверяем двухсимвольные сочетания
			pair := string(runes[i]) + string(runes[i+1])
			if replacement, ok := translitMap[pair]; ok {
				result.WriteString(replacement)
				i += 2
				continue
			}
		}

		// Проверяем одиночный символ
		char := string(runes[i])
		if replacement, ok := translitMap[char]; ok {
			result.WriteString(replacement)
		} else {
			result.WriteRune(runes[i])
		}
		i++
	}

	return result.String()
}
