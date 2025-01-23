package functions

import (
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func get_file_excel() *excelize.File {
	nameFileSlice, _ := getExcelName()
	nameFile := nameFileSlice[0]
	fileLocate := fmt.Sprintf("uploaded_files/%s", nameFile)
	f, err := excelize.OpenFile(fileLocate)
	if err != nil {
		fmt.Println(err)
	}
	return f
} // Файл Эксель
func get_sheets(f *excelize.File) []string {
	var all_sheets []string
	for i := 1; i <= f.SheetCount; i++ {
		all_sheets = append(all_sheets, f.GetSheetName(i))
	}
	return all_sheets
}
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
