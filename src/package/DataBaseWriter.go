package funcExcel

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

func FunctionDbWriter() {
	fmt.Println("Functions 7")
	schedule()
}

func schedule() {

	connStr := "user=postgres password=password sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к серверу PostgreSQL: ", err)
	}
	defer db.Close()

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		log.Fatal("Не удалось подключиться к серверу PostgreSQL:", err)
	}
	fmt.Println("Успешно подключено к серверу PostgreSQL!")

	// ——————————————————————————————————————————————————————————————————————————————————————
	f := get_file_excel()
	sheets := get_sheets(f)

	for index, sheet := range sheets {
		if checkString(string(sheet[0])) {
			sheet = string(sheet[1:]) + string(sheet[0])
		}
		dbName := strings.ToLower(replaceCyrillicWithLatin(removeSpaces(removeDots(sheet))))

		_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			fmt.Printf("Ошибка при создании базы данных %s: %s\n", dbName, err)
		} else {
			fmt.Printf("База данных %s успешно создана!\n", dbName)
		}

		groupsList := get_all_groups(f, sheets[index])
		// pre_psGroups := make([]string, 0)
		for _, groups := range groupsList {
			group := groups[0]
			group = strings.ToLower(replaceCyrillicWithLatin(removeSpaces(group)))
			connStr := fmt.Sprintf("user=postgres password=password dbname=%s sslmode=disable", dbName)
			courseDB, err := sql.Open("postgres", connStr)
			if err != nil {
				log.Printf("Ошибка подключения к базе данных курса: %s", err)
			}
			defer courseDB.Close()
			tableName := fmt.Sprintf(`"%s"`, group)

			// pre_psGroups = append(pre_psGroups, group)

			createTableSQL := fmt.Sprintf(`
				CREATE TABLE %s (
					day_of_week TEXT,
					time TEXT,
					subject TEXT,
					auditory TEXT,
					teacher TEXT,
					weeks TEXT
				)`, tableName)

			_, err = courseDB.Exec(createTableSQL)
			if err != nil {
				fmt.Printf("Ошибка при создании таблицы для группы %s: %s\n", replaceCyrillicWithLatin(group), err)
			} else {
				fmt.Printf("Таблица для группы %s успешно создана!\n", replaceCyrillicWithLatin(group))
			}

			seen := make([]string, 0)
			for i := 1; i < len(groups); i++ {
				cellGroup := string(groups[i])
				mainCell := string(cellGroup[0])
				startRow, err := strconv.Atoi(string(cellGroup[1]))
				if err != nil {
					fmt.Println("Ошибка преобразования:", err)
				} else {
					startRow += 1
				}

				for k := startRow; k < len(f.GetRows(sheets[index]))+1; k++ {
					day := removeExtraSpaces(f.GetCellValue(sheets[index], fmt.Sprintf("%s%d", "A", k)))
					time := removeExtraSpaces(f.GetCellValue(sheets[index], fmt.Sprintf("%s%d", "B", k)))
					couple := removeExtraSpaces(f.GetCellValue(sheets[index], fmt.Sprintf("%s%d", string(mainCell), k)))

					// Если значение в ячейке не пустое
					if couple != "" {
						// Формируем уникальный ключ для комбинации day, time, couple
						key := fmt.Sprintf("%s|%s|%s", day, time, couple)

						// Проверяем, существует ли такая комбинация в карте
						if !contains(seen, key) {
							// Если комбинации нет, добавляем её в карту
							seen = append(seen, key)

							// Дальше обрабатываем пару (day, time, couple) как уникальную
							classCouple := parseClassInfo(couple)

							// Ваш дальнейший код для обработки уникальной пары
							// Например:
							// couples = append(couples, classCouple)
							scheduleGroup := []struct {
								dayWeek    string
								timeCouple string
								subject    string
								auditory   string
								teacher    string
								weeks      string
							}{
								{day, time, classCouple.Subject, classCouple.Auditory, classCouple.Teacher, classCouple.Weeks},
							}

							for _, lesson := range scheduleGroup {
								_, err := courseDB.Exec(fmt.Sprintf(`
								INSERT INTO %s (day_of_week, time, subject, auditory, teacher, weeks)
								SELECT $1, $2, $3, $4, $5, $6
								WHERE NOT EXISTS (
									SELECT 1 FROM %s WHERE day_of_week = $1 AND time = $2 AND subject = $3 AND auditory = $4 AND teacher = $5 AND weeks = $6 
									)
								`, tableName, tableName), lesson.dayWeek, lesson.timeCouple, lesson.subject, lesson.auditory, lesson.teacher, lesson.weeks)
								if err != nil {
									fmt.Printf("Ошибка при добавлении записи в таблицу для группы %s: %s\n", group, err)
									fmt.Println(tableName, scheduleGroup)
								} else {
									fmt.Printf("Запись для группы %s успешно добавлена!\n", group)

								}
							}
						}
					}
				}
			}
		}
	}
}
