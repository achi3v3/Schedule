package funcExcel

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

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
