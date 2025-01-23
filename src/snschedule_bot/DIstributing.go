package functions

import (
	"regexp"
	"strings"
)

// ===========================================MAIN==============================================================
func parseClassInfo(input string) ClassInfo {
	input = strings.TrimSpace(input)
	// inputWithoutBugs :=

	auditory := parseAuditory(input)
	weeks := parseWeeks(input)
	teacher := parseTeacher(input)
	subject := parseSubject(input)
	if (removeSpaces(weeks) == "" || removeSpaces(auditory) == "") && (strings.Contains(subject, "(") && strings.Contains(subject, ")")) {
		l := strings.Index(subject, "(")
		r := strings.Index(subject, ")")
		noneObject := removeExtraSpaces(subject[l+1 : r])
		if removeSpaces(auditory) == "" && checkString(noneObject) {
			if len(noneObject) > 2 {
				auditory = subject[l : r+1]
				subject = strings.Replace(subject, auditory, "", -1)
			}
		} else if removeSpaces(weeks) == "" {
			weeks = subject[l : r+1]
			subject = strings.Replace(subject, weeks, "", -1)
		}
	}
	if auditory == "" {
		auditory = "—"
	}
	if weeks == "" {
		weeks = "—"
	}
	if teacher == "" {
		teacher = "—"
	}
	return ClassInfo{
		Subject:  strings.TrimSpace(subject),
		Auditory: auditory,
		Teacher:  teacher,
		Weeks:    weeks,
	}
}

// ===========================================SUBJECT==============================================================
func parseSubject(input string) string {
	auditory := parseAuditory(input)
	weeks := parseWeeks(input)
	teacher := parseTeacher(input)

	input = strings.TrimSpace(input)
	input = strings.Replace(input, auditory, "", -1)
	input = strings.Replace(input, teacher, "", -1)
	input = strings.Replace(input, weeks, "", -1)

	if strings.Contains(weeks, "нед") {
		input = strings.Replace(input, weeks, "", -1)
	}
	subject := strings.TrimSpace(input)
	return removeDuplicates(subject)
}

// ===========================================AUDITORY==============================================================
func parseAuditory(input string) string {
	auditoryRe := regexp.MustCompile(`\(\s*(ауд(итор(ия)?)?\.?[\s,]*[\wа-яА-Я]+)\s*\)`)
	matches := auditoryRe.FindString(input)
	if removeSpaces(matches) == "" {
		if strings.Contains(input, "ауд") {
			aud := strings.Index(input, "ауд")

			if strings.Contains(input, "(") && strings.Contains(input, ")") {
				// Ищем первую пару скобок
				l := strings.Index(input, "(")
				r := strings.Index(input, ")")

				// Проверяем, что "ауд" находится внутри этих скобок
				if l <= aud && r >= aud {
					matches = input[l : r+1]
				} else {
					// Если "ауд" не в первой паре скобок, ищем последнюю пару
					l = strings.LastIndex(input, "(")
					r = strings.LastIndex(input, ")") // Исправлено: ищем последнюю правую скобку

					if l <= aud && r >= aud {
						matches = input[l : r+1]
					}
				}
			}
		}
	}
	if removeSpaces(matches) == "" {
		if strings.Contains(strings.ToUpper(input), "ДИСТАНТ") {
			re := regexp.MustCompile(`\(?ДИСТАНТ\)?`)
			matches = re.FindString(input)
		}
	}
	return strings.TrimSpace(matches)
}

// ===========================================TEACHER==============================================================
func parseTeacher(input string) string {
	input = strings.Replace(input, ". Практикум", " Практикум", -1)
	input = strings.Replace(input, parseAuditory(input), "", -1)
	input = strings.Replace(input, parseWeeks(input), "", -1)

	teacherRe := regexp.MustCompile(`(доц\.|проф\.|ст\.преп\.|асс\.|преп\.)\s+[A-Za-zА-Яа-яёЁ\.]+\s+[A-Za-zА-Яа-яёЁ\.]+`)

	matches := teacherRe.FindString(input)

	if removeSpaces(matches) == "" {
		titles := []string{"доц.", "проф.", "ст. преп.", "ст. преп", "асс.", "преп."}

		for _, title := range titles {
			if strings.Contains(input, title) {
				titleIndex := strings.Index(input, title)
				substr := input[titleIndex+len(title):]

				teacherReAlt := regexp.MustCompile(`[A-Za-zА-Яа-яёЁ\.]+\s+[A-Za-zА-Яа-яёЁ\.]+`)

				matches = teacherReAlt.FindString(substr)

				if matches != "" {
					return strings.TrimSpace(title + " " + matches)
				}
			}
		}
	}
	if removeSpaces(matches) == "" {
		familyNameRe := regexp.MustCompile(`[A-Za-zА-Яа-яёЁ]+(\s*[A-Za-zА-Яа-яёЁ]\.\s*)+`)
		matches = familyNameRe.FindString(input)

		if matches != "" {
			return strings.TrimSpace(matches)
		}
	}
	if removeSpaces(matches) == "" {
		teacherRe := regexp.MustCompile(`(доц\.|проф\.|ст\.преп\.|асс\.|преп\.)\s+[A-Za-zА-Яа-яёЁ]+\s+[A-ZА-Я]\.[A-ZА-Я]\.`)

		matches := teacherRe.FindString(input)
		if matches != "" {
			return strings.TrimSpace(matches)
		}

		initialsAndNameRe := regexp.MustCompile(`([A-Za-zА-Яа-яёЁ]+)\s([A-ZА-Я]\.[A-ZА-Я]\.)`)

		matches = initialsAndNameRe.FindString(input)
		if matches != "" {
			return strings.TrimSpace(matches)
		}

		familyNameRe := regexp.MustCompile(`[A-Za-zА-Яа-яёЁ]+(\s+[A-ZА-Я]\.)+`)

		matches = familyNameRe.FindString(input)
		if matches != "" {
			return strings.TrimSpace(matches)
		}
	}

	return strings.TrimSpace(matches)
}

// ===========================================WEEKS==============================================================
func parseWeeks(input string) string {
	weeksRe := regexp.MustCompile(`\(\s*(нед(ели|еля)?\.?|\d+-\d+|\S*нед\S*\s*\d+-\d+)\s*\)`)
	matches := weeksRe.FindString(input)
	if removeSpaces(matches) == "" {
		if strings.Contains(input, "нед") {
			aud := strings.Index(input, "нед")
			if strings.Contains(input, "(") && strings.Contains(input, ")") {
				l := strings.Index(input, "(")
				r := strings.Index(input, ")")

				if l <= aud && r >= aud {
					matches = input[l : r+1]
				} else {
					l = strings.LastIndex(input, "(")
					r = strings.LastIndex(input, ")")

					if l <= aud && r >= aud {
						matches = input[l : r+1]
					}
				}
			}
		}
	}
	return strings.TrimSpace(matches)
}

// ===========================================ADDONS=REMOVERS==============================================================
func removeDuplicates(input string) string {
	words := strings.Fields(input)
	seen := make(map[string]bool)
	var result []string

	for _, word := range words {
		if !seen[word] {
			seen[word] = true
			result = append(result, word)
		}
	}

	return strings.Join(result, " ")
}
func removeBrackets(text string) string {
	text = strings.ReplaceAll(text, "(", "")
	text = strings.ReplaceAll(text, ")", "")
	return text
}

// ===========================================STRUCT==============================================================
type ClassInfo struct {
	Subject  string
	Auditory string
	Teacher  string
	Weeks    string
}
