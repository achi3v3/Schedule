package funcExcel

import (
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func FunctionExcelReader() {
	fmt.Println("func: Excel Reader")

}

// ================================================GET=FUNCTIONS==============================================================================

func getCouples(f *excelize.File, sheet string, group []string) [][]string {
	daysWeek := []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}

	cellDays := make([][]string, len(daysWeek))
	couplesCellDays := make([][]string, len(daysWeek))

	for col := 1; col < 3; col++ {
		for row := 1; row < len(f.GetRows(sheet)); row++ {
			if contains(daysWeek, f.GetCellValue(sheet, fmt.Sprintf("%s%d", excelize.ToAlphaString(col-1), row))) {
				ind := findIndex(daysWeek, removeExtraSpaces(f.GetCellValue(sheet, fmt.Sprintf("%s%d", excelize.ToAlphaString(col-1), row))))
				cellDays[ind] = append(cellDays[ind], fmt.Sprintf("%s%d", excelize.ToAlphaString(col-1), row))
			}
		}
	}
	seen := make(map[string]bool)
	for i := 1; i < len(group); i++ {
		cell1 := group[i][:1]
		for j := 0; j < len(cellDays); j++ {
			for k := 0; k < len(cellDays[j]); k++ {
				coupleData := (f.GetCellValue(sheet, cell1+cellDays[j][k][1:]))
				if removeExtraSpaces(coupleData) != "" {
					data := fmt.Sprintf("%s # %s",
						removeExtraSpaces(f.GetCellValue(sheet, "B"+cellDays[j][k][1:])),
						removeExtraSpaces(f.GetCellValue(sheet, cell1+cellDays[j][k][1:])))

					// Используем map для проверки, был ли уже добавлен этот элемент
					if _, exists := seen[data]; !exists {
						// Если еще нет в map, добавляем в срез и в map
						seen[data] = true
						couplesCellDays[j] = append(couplesCellDays[j], data)
					}
				}
			}
		}
	}

	return couplesCellDays
}
