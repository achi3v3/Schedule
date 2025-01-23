package functions

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func createTableUsers(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL UNIQUE,
		username TEXT NOT NULL,
		first_name TEXT NOT NULL,
		registration_time TIMESTAMP NOT NULL,
		user_course TEXT DEFAULT '',
		user_group TEXT DEFAULT ''
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Ошибка при создании таблицы: %s", err)
	}
	fmt.Println("USERS OKAY")
}
func addUser(userID int64, username, firstName string) bool {
	connStr := "user=postgres password=password sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Ошибка при подключении к базе данных: %s", err)
	}
	defer db.Close()
	var exists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`
	err = db.QueryRow(queryCheck, userID).Scan(&exists)

	if err != nil {
		fmt.Println(err)
		return false
	}
	if exists {
		fmt.Println(exists)
		return false
	}
	queryInsert := `
		INSERT INTO users (user_id, username, first_name, registration_time, user_course, user_group)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	registrationTime := time.Now()
	userGroup := ""
	userCourse := ""
	_, err = db.Exec(queryInsert, userID, username, firstName, registrationTime, userCourse, userGroup)
	if err != nil {
		return false
	}
	fmt.Println("Пользователь успешно добавлен в базу данных!")
	return true
}
func createDataBasesExcel(db *sql.DB) {
	createTableSheets(db)
	createTableGroups(db)

	sheets := get_sheets(get_file_excel())
	for _, sheet := range sheets {
		addSheets(db, sheet)
		groups := get_groups(get_file_excel(), sheet)
		for _, group := range groups {
			addGroups(db, sheet, group)
		}
	}
}
func createTableSheets(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS sheets (
		id SERIAL PRIMARY KEY,
		course TEXT NOT NULL UNIQUE
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Ошибка при создании таблицы: %s", err)
	}
	fmt.Println("TABLESHEETS OKAY")
}
func createTableGroups(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS groups (
		id SERIAL PRIMARY KEY,
		sheet TEXT NOT NULL,
		sheetRu TEXT NOT NULL,
		namegroup TEXT NOT NULL UNIQUE,
		namegroupRu TEXT NOT NULL UNIQUE
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Ошибка при создании таблицы: %s", err)
	}
	fmt.Println("TABLEGROUPS OKAY")
}
func addGroups(db *sql.DB, sheet, group string) error {
	var exists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM groups WHERE namegroup = $1)`
	err := db.QueryRow(queryCheck, group).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования namegroup: %v", err)
	}
	if exists {
		fmt.Println("namegroup с таким названием уже существует")
		return nil
	}
	queryInsert := `
		INSERT INTO groups (sheet, sheetRu, namegroup, namegroupRu)
		VALUES ($1, $2, $3, $4)
	`
	_, err = db.Exec(queryInsert, renameSheetGroup(sheet), sheet, renameSheetGroup(group), group)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении namegroup: %v", err)
	}

	fmt.Println("namegroup успешно добавлен в базу данных!")
	return nil
}
func addSheets(db *sql.DB, sheet string) error {
	var exists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM sheets WHERE course = $1)`
	err := db.QueryRow(queryCheck, sheet).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования пользователя: %v", err)
	}
	if exists {
		fmt.Println("Course с таким названием уже существует")
		return nil
	}
	queryInsert := `
        INSERT INTO sheets (course)
        VALUES ($1)
    `
	_, err = db.Exec(queryInsert, sheet)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении course: %v", err)
	}

	fmt.Println("Course успешно добавлен в базу данных!")
	return nil
}
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

func createFilesTable(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS files (
		id SERIAL PRIMARY KEY,
		file_name VARCHAR(255) NOT NULL,
		week_number TEXT
	);
	`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("ошибка при создании таблицы: %v", err)
	}

	fmt.Println("Таблица 'files' успешно создана")
	return nil
}
func PinGroup(ctx context.Context, userID int64, course, group string) (string, error) {
	connStr := "user=postgres password=password sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return "", fmt.Errorf("ошибка при подключении к базе данных: %w", err)
	}
	defer db.Close()

	var exists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`
	err = db.QueryRowContext(ctx, queryCheck, userID).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf("ошибка при проверке пользователя: %w", err)
	}

	if !exists {
		return "Пользователь не найден в системе.", nil
	}

	var currentCourse, currentGroup string
	querySelect := `SELECT user_course, user_group FROM users WHERE user_id = $1`
	err = db.QueryRowContext(ctx, querySelect, userID).Scan(&currentCourse, &currentGroup)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении текущих данных пользователя: %w", err)
	}

	if currentCourse == course && currentGroup == group {
		return fmt.Sprintf("🆘 Группа уже закреплена:\n<blockquote>Уровень обучения: %s\nГруппа: %s</blockquote>", course, group), nil
	}

	queryUpdate := `
		UPDATE users
		SET user_course = $1, user_group = $2
		WHERE user_id = $3
	`
	_, err = db.ExecContext(ctx, queryUpdate, course, group, userID)
	if err != nil {
		return "", fmt.Errorf("ошибка при обновлении данных пользователя: %w", err)
	}

	return fmt.Sprintf("✅ Данные успешно закреплены:\n<blockquote>Уровень обучения: %s\nГруппа: %s</blockquote>", course, group), nil
}
func GetUserCourseAndGroup(ctx context.Context, userID int64) (string, string, error) {
	connStr := "user=postgres password=password sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при подключении к базе данных: %w", err)
	}
	defer db.Close()

	var course, group string
	query := `SELECT user_course, user_group FROM users WHERE user_id = $1`
	err = db.QueryRowContext(ctx, query, userID).Scan(&course, &group)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("пользователь с ID %d не найден", userID)
		}
		return "", "", fmt.Errorf("ошибка при извлечении данных: %w", err)
	}

	return course, group, nil
}
func getExcelName() ([]string, error) {
	tableName := "files"
	connStr := "user=postgres password=password sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Ошибка подключения к серверу PostgreSQL: %s", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("Не удалось подключиться к серверу PostgreSQL: %s", err)
	}
	fmt.Println("Успешно подключено к серверу PostgreSQL!")

	query := fmt.Sprintf("SELECT file_name, week_number FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer rows.Close()

	var dataFiles []string
	for rows.Next() {
		var fileName string
		var weekNumber int

		err := rows.Scan(&fileName, &weekNumber)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки: %w", err)
		}
		dataFiles = append(dataFiles, fileName)
		dataFiles = append(dataFiles, fmt.Sprintf("%d", weekNumber))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при завершении перебора строк: %w", err)
	}

	return dataFiles, nil
}
func getAllSheets() ([]string, error) {

	connStr := "user=postgres password=password sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Ошибка при подключении к базе данных: %s", err)
	}
	defer db.Close()

	var sheets []string

	rows, err := db.Query("SELECT course FROM sheets")
	if err != nil {
		return nil, fmt.Errorf("ошибка при извлечении курсов: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var course string
		if err := rows.Scan(&course); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных: %s", err)
		}
		sheets = append(sheets, course)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результата запроса: %s", err)
	}

	return sheets, nil
}
func FunctionDbWriter() {
	fmt.Println("Function DB Writer")
	schedule()
}
func ReloadFile(fileName string, week string) {
	connStr := "user=postgres password=password sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("DataBaseNames: Ошибка подключения к серверу PostgreSQL: %s", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("DataBaseNames: Не удалось подключиться к серверу PostgreSQL: %s", err)
	}

	fmt.Println("DataBaseNames: Успешно подключено к серверу PostgreSQL!")

	err = db.Ping()
	if err != nil {
		log.Printf("DataBaseNames: Не удалось подключиться к серверу PostgreSQL: %s", err)
	}
	createFilesTable(db)
	deleteAllFromTable(db, "files")

	insertFileData(db, fileName, week)
	deleteAllSchedule()
	FunctionDbWriter()
}
func deleteAllFromTable(db *sql.DB, tableName string) error {
	query := fmt.Sprintf("DELETE FROM %s;", tableName)

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("ошибка при удалении данных: %v", err)
	}

	fmt.Println("Все данные успешно удалены из таблицы", tableName)
	return nil
}
func insertFileData(db *sql.DB, fileName string, weekNumber string) error {
	insertQuery := `
		INSERT INTO files (file_name, week_number)
		VALUES ($1, $2);
	`
	_, err := db.Exec(insertQuery, fileName, weekNumber)
	if err != nil {
		return fmt.Errorf("ошибка при вставке данных: %v", err)
	}

	fmt.Println("Данные успешно добавлены")
	return nil
}
func deleteAllSchedule() {
	allDataBases := FunctionDataBaseNames()
	for _, database := range allDataBases {
		connStr := fmt.Sprintf("user=postgres password=password dbname=%s sslmode=disable", database)
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Ошибка при подключении к базе данных: %s", err)
		}
		defer db.Close()
		allTables := FunctionTableNames(database)
		for _, table := range allTables {
			if table != "users" && table != "files" {
				deleteAllFromTable(db, table)
				fmt.Printf("Все данные из таблицы %s удалены\n", table)
			}
		}
	}
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
func FunctionDataBaseNames() []string {
	connStr := "user=postgres password=password sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("DataBaseNames: Ошибка подключения к серверу PostgreSQL: %s", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("DataBaseNames: Не удалось подключиться к серверу PostgreSQL: %s", err)
	}

	fmt.Println("DataBaseNames: Успешно подключено к серверу PostgreSQL!")

	err = db.Ping()
	if err != nil {
		log.Printf("DataBaseNames: Не удалось подключиться к серверу PostgreSQL: %s", err)
	}

	return RequestCourses(db)
} // Базы Данных
func FunctionTableNames(group string) []string {
	tbName := strings.ToLower(replaceCyrillicWithLatin(removeSpaces(group)))

	connStr := "user=postgres password=password sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("TableNames: Ошибка подключения к серверу PostgreSQL: %s", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("TableNames: Не удалось подключиться к серверу PostgreSQL: %s", err)
	}

	fmt.Println("TableNames: Успешно подключено к серверу PostgreSQL!")

	return RequestGroupes(db, tbName)
} // Таблицы
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
func FunctionDataBaseReader() {
	connStr := "user=postgres password=password sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("DataBaseReader: Ошибка подключения к серверу PostgreSQL: %s", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("DataBaseReader: Не удалось подключиться к серверу PostgreSQL: %s", err)
	}

	fmt.Println("DataBaseReader: Успешно подключено к серверу PostgreSQL!")
}
func RequestCourses(db *sql.DB) []string {
	query := `SELECT datname FROM pg_database WHERE datistemplate = false;`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса: %s", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			log.Printf("Ошибка при чтении результата запроса: %s", err)
		}
		databases = append(databases, dbName)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при обходе строк: %s", err)
	}
	return databases
} // ПОлучаем БДшки
func ConnectDB(dbName string) (db *sql.DB) {
	connStr := fmt.Sprintf("user=postgres password=password dbname=%s sslmode=disable", dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Не удалось подключиться к базе данных %s: %s\n", dbName, err)
	}
	defer db.Close()
	return db
}
func RequestGroupes(db *sql.DB, dbName string) []string {
	connStr := fmt.Sprintf("user=postgres password=password dbname=%s sslmode=disable", dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Не удалось подключиться к базе данных %s: %s\n", dbName, err)
	}
	defer db.Close()

	tablesQuery := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`
	tablesRows, err := db.Query(tablesQuery)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса для базы данных %s: %s\n", dbName, err)
	}
	defer tablesRows.Close()

	var tables []string
	for tablesRows.Next() {
		var tableName string
		if err := tablesRows.Scan(&tableName); err != nil {
			log.Printf("Ошибка при чтении результата запроса для базы данных %s: %s\n", dbName, err)
			continue
		}
		tables = append(tables, tableName)
	}

	if err := tablesRows.Err(); err != nil {
		log.Printf("Ошибка при обходе строк таблиц для базы данных %s: %s\n", dbName, err)
	}
	return tables
} // Получаем Таблицы
type Couple struct {
	day      string
	time     string
	name     string
	auditory string
	teacher  string
	weeks    string
} // структура для расписания
func RequestGroupDay(db *sql.DB, dbName, table string, day string) [][]string {

	connStr := fmt.Sprintf("user=postgres password=password dbname=%s sslmode=disable", dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Не удалось подключиться к базе данных %s: %s\n", dbName, err)
	}
	defer db.Close()

	dataCouple, err := getDataFromTable(db, table)
	if err != nil {
		log.Printf("Ошибка при получении данных из таблицы: %s", err)
	}
	dataCouple = sortCouplesByTime(dataCouple)
	var coupleList [][]string
	for _, couple := range dataCouple {
		if couple.day == day {
			coupleList = append(coupleList, []string{couple.day, couple.time, couple.name, couple.auditory, couple.teacher, couple.weeks})
		}
	}
	return coupleList
}
func getDataFromTable(db *sql.DB, tableName string) ([]Couple, error) {
	query := fmt.Sprintf("SELECT day_of_week, time, subject, auditory, teacher, weeks FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer rows.Close()

	var DataCouples []Couple
	for rows.Next() {
		var DataCouple Couple
		if err := rows.Scan(&DataCouple.day, &DataCouple.time, &DataCouple.name, &DataCouple.auditory, &DataCouple.teacher, &DataCouple.weeks); err != nil {
			return nil, fmt.Errorf("ошибка при чтении строки: %w", err)
		}
		DataCouples = append(DataCouples, DataCouple)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при переборе строк: %w", err)
	}

	return DataCouples, nil
} // Получаем данные с таблиц
func extractStartTime(timeStr string) (float64, error) {
	parts := strings.Split(timeStr, "-")
	if len(parts) < 2 {
		return 0, fmt.Errorf("неправильный формат времени: %s", timeStr)
	}
	startTimeStr := parts[0]
	startTime, err := strconv.ParseFloat(startTimeStr, 64)
	if err != nil {
		return 0, fmt.Errorf("ошибка при преобразовании времени: %v", err)
	}

	return startTime, nil
} // помощь в сортировке
func sortCouplesByTime(couples []Couple) []Couple {
	sort.Slice(couples, func(i, j int) bool {
		startTimeI, _ := extractStartTime(couples[i].time)
		startTimeJ, _ := extractStartTime(couples[j].time)
		return startTimeI < startTimeJ
	})

	return couples
} // Функция для сортировки слайса Couple по времени
func getGroupsByCourseRu(courseName string) ([]string, error) {
	connStr := "user=postgres password=password sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Ошибка при подключении к баз %sе данных: ", err)
	}
	defer db.Close()

	var groups []string

	rows, err := db.Query("SELECT namegroupRu FROM groups WHERE sheet = $1", renameSheetGroup(courseName))
	if err != nil {
		return nil, fmt.Errorf("ошибка при извлечении групп для курса '%s': %v", courseName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var group string
		if err := rows.Scan(&group); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных: %v", err)
		}
		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результата запроса: %v", err)
	}

	return groups, nil
}
func FunctionDataBaseTableData(dbName, group, day string) [][]string {

	dbName = (dbName)
	tbName := (group)
	connStr := "user=postgres password=password sslmode=disable"

	// Открытие соединения с PostgreSQL сервером
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("DTBC: Ошибка подключения к серверу PostgreSQL: %s", err)
	}
	defer db.Close()

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		log.Printf("DTBC: Не удалось подключиться к серверу PostgreSQL: %s", err)
	}

	return RequestGroupDay(db, dbName, tbName, day)
} // Данные Расписания
