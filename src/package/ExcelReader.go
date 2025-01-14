package funcExcel

import (
	"fmt"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func FunctionExcelReader() {
	fmt.Println("func: Excel Reader")
}

// ================================================GET=FUNCTIONS==============================================================================
func get_file_excel() *excelize.File {
	nameFileSlice, _ := getExcelName()
	nameFile := nameFileSlice[0]
	// nameFile := "6625409928_bit_tzi201_17.xlsx"
	fileLocate := fmt.Sprintf("uploaded_files/%s", nameFile)
	f, err := excelize.OpenFile(fileLocate)
	if err != nil {
		fmt.Println(err)
	}
	return f
} // Файл Эксель
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

	for i := 1; i < len(group); i++ {
		cell1 := group[i][:1]
		for j := 0; j < len(cellDays); j++ {
			for k := 0; k < len(cellDays[j]); k++ {
				coupleData := (f.GetCellValue(sheet, cell1+cellDays[j][k][1:]))
				if removeExtraSpaces(coupleData) != "" {
					data := fmt.Sprintf("%s # %s", removeExtraSpaces(f.GetCellValue(sheet, "B"+cellDays[j][k][1:])), removeExtraSpaces(f.GetCellValue(sheet, cell1+cellDays[j][k][1:])))
					couplesCellDays[j] = append(couplesCellDays[j], data)
				}
			}
		}
	}

	return couplesCellDays
}
func get_all_groups(f *excelize.File, sheet string) [][]string {
	var col, row int
	var result [][]string
	row, col = 1, 1
	rows := len(f.GetRows(sheet))
	flag_row := false

	for i := 0; i < get_len_sheet(f, sheet); i++ {
		if !flag_row {
			for row = 0; row < rows; row++ {
				group := removeExtraSpaces(f.GetCellValue(sheet, fmt.Sprintf("%s%d", excelize.ToAlphaString(col-1), row)))
				cell_of_group := fmt.Sprintf("%s%d", excelize.ToAlphaString(col-1), row)
				if len(strings.Fields(group)) == 2 {
					if isValidFormat(strings.Fields(group)[0]) {
						if !containsInNested(result, group) {
							result = append(result, []string{group, cell_of_group})
						} else {
							findAndAdd(result, group, cell_of_group)
						}
					}
				} else if isValidFormat((group)) {
					if !containsInNested(result, group) {
						result = append(result, []string{group, cell_of_group})
					} else {
						findAndAdd(result, group, cell_of_group)
					}
				}
			}
		}
		col += 1
	}
	return result
} // ГРУППЫ С ИХ ЯЧЕЙКАМИ
func get_groups(f *excelize.File, sheet string) []string {
	var col, row int
	var result []string
	row, col = 1, 1
	rows := len(f.GetRows(sheet))
	flag_row := false

	for i := 0; i < get_len_sheet(f, sheet); i++ {
		if !flag_row {
			for row = 0; row < rows; row++ {
				group := removeExtraSpaces(f.GetCellValue(sheet, fmt.Sprintf("%s%d", excelize.ToAlphaString(col-1), row)))
				if len(strings.Fields(group)) == 2 {
					if isValidFormat(strings.Fields(group)[0]) {
						if !contains(result, group) {
							result = append(result, group)
						}
					}
				} else if isValidFormat((group)) {
					if !contains(result, group) {
						result = append(result, group)
					}
				}
			}
		}
		col += 1
	}
	return result
} // ГРУППЫ БЕЗ ЯЧЕЕК
func get_len_sheet(f *excelize.File, sheet string) int {
	var col, row, counter, checker int
	row = 1
	rows := len(f.GetRows(sheet))

	for col = 1; ; col++ {
		if row >= rows && checker > 2 {
			counter = (col - 1) - checker
			break
		}
		if f.GetCellValue(sheet, fmt.Sprintf("%s%d", excelize.ToAlphaString(col-1), row)) == "" {
			for row = 1; ; row++ {
				if row >= rows {
					checker += 1
					break
				}
				if f.GetCellValue(sheet, fmt.Sprintf("%s%d", excelize.ToAlphaString(col-1), row)) != "" {
					col += 1
					row = 1
					checker = 0
					break
				}
			}
		}
	}
	return counter
}
func get_sheets(f *excelize.File) []string {
	var all_sheets []string
	for i := 1; i <= f.SheetCount; i++ {
		all_sheets = append(all_sheets, f.GetSheetName(i))
	}
	return all_sheets
}
