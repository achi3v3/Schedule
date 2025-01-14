package funcExcel

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

func FunctionDataBaseFunctions() {
	fmt.Println("func: Data Base Functions")
}

func FunctionDataBaseTableData(dbName, group, day string) [][]string {

	dbName = renameSheetGroup(dbName)
	tbName := renameSheetGroup(group)
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
func ReloadFile(fileName string) {
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
	insertFileData(db, fileName, 20)
	deleteAllSchedule(db)
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
func createFilesTable(db *sql.DB) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS files (
		id SERIAL PRIMARY KEY,
		file_name VARCHAR(255) NOT NULL,
		week_number INT
	);
	`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("ошибка при создании таблицы: %v", err)
	}

	fmt.Println("Таблица 'files' успешно создана")
	return nil
}
func insertFileData(db *sql.DB, fileName string, weekNumber int) error {
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
func deleteAllSchedule(db *sql.DB) {
	allDataBases := FunctionDataBaseNames()
	for _, database := range allDataBases {
		allTables := FunctionTableNames(database)
		for _, table := range allTables {
			if table != "users" && table != "files" {
				deleteAllFromTable(db, table)
			}
		}
	}
}
