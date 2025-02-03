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
	// "github.com/360EntSecGroup-Skylar/excelize"
)

type Couple struct {
	day      string
	time     string
	name     string
	auditory string
	teacher  string
	weeks    string
	added    bool
	student  string
} // структура для расписания
func createTableUsers(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL UNIQUE,
		username TEXT NOT NULL,
		first_name TEXT NOT NULL,
		registration_time TIMESTAMP NOT NULL,
		user_course TEXT DEFAULT '',
		user_group TEXT DEFAULT '',
		redactor BOOLEAN DEFAULT FALSE,
		admin BOOLEAN DEFAULT FALSE,
		permcourse TEXT DEFAULT '',
		permgroup TEXT DEFAULT ''
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("%s %s", errorCreateTable, err)
	}
	fmt.Println("✅ ok - createTableUsers")
}

const (
	user     = "postgres"
	password = "password"
	sslmode  = "disable"
	host     = "localhost"
	port     = 5432

	errorConnect     = "Ошибка при подключении к базе данных."
	errorCreateTable = "Ошибка при создании таблицы. "

	errorSendMsg  = "Ошибка отправки сообщения:"
	errorSendFile = "Не удалось отправить файл:"

	joinUser = "Пользователь успешно добавлен в базу данных!"
	postgres = "postgres"

	notExist = "уже существует\n"
)

func addUser(userID int64, username, firstName string) bool {
	connStr := fmt.Sprintf("user=%s password=%s sslmode=%s", user, password, sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("%s %s", errorConnect, err)
	}
	defer db.Close()

	var exists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`
	err = db.QueryRow(queryCheck, userID).Scan(&exists)

	if err != nil {
		return false
	}
	if exists {
		return false
	}
	queryInsert := `
		INSERT INTO users (user_id, username, first_name, registration_time, user_course, user_group)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	registrationTime := time.Now()
	userGroup, userCourse := "", ""

	_, err = db.Exec(queryInsert, userID, username, firstName, registrationTime, userCourse, userGroup)
	if err != nil {
		return false
	}

	fmt.Println(joinUser)
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
		log.Printf("%s %s", errorCreateTable, err)
	}
	fmt.Println("✅ ok - createTableSheets")
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
		log.Printf("%s %s", errorCreateTable, err)
	}
	fmt.Println("✅ ok - createTableGroups")
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

	fmt.Printf("✅ ok - %s успешно добавлен в базу данных!", group)
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
		fmt.Printf("✅ ok - %s %s", sheet, notExist)
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

	fmt.Printf("✅ ok - %s успешно добавлен в базу данных!", sheet)
	return nil
}
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

	fmt.Println("✅ ok - 'files' таблица успешно создана")
	return nil
}

func Connect(dbName string) (*sql.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, password, host, port, dbName, sslmode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("❌ Не удалось подключиться к базе данных %s: %v\n", dbName, err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Printf("❌ База данных %s недоступна: %v\n", dbName, err)
		db.Close()
		return nil, err
	}

	return db, nil
}
func RequestGroupDay(dbName, table string, day string) [][]string {
	db, err := Connect(dbName)
	if err != nil {
		log.Println("RequestGroupDay:", err)
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
			if !couple.added {
				coupleList = append(coupleList, []string{couple.day, couple.time, couple.name, couple.auditory, couple.teacher, couple.weeks})
			} else {
				coupleList = append(coupleList, []string{couple.day, couple.time, couple.name, couple.auditory, couple.teacher, couple.weeks, couple.student})
			}
		}
	}
	return coupleList
}
func RequestGroupes(dbName string) []string {
	db, err := Connect(dbName)
	if err != nil {
		log.Println("RequestGroupes:", err)
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
func getDataFromTable(db *sql.DB, tableName string) ([]Couple, error) {
	query := fmt.Sprintf("SELECT day_of_week, time, subject, auditory, teacher, weeks, added, student FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer rows.Close()

	var DataCouples []Couple
	for rows.Next() {
		var DataCouple Couple
		if err := rows.Scan(&DataCouple.day, &DataCouple.time, &DataCouple.name, &DataCouple.auditory, &DataCouple.teacher, &DataCouple.weeks, &DataCouple.added, &DataCouple.student); err != nil {
			return nil, fmt.Errorf("ошибка при чтении строки: %w", err)
		}
		DataCouples = append(DataCouples, DataCouple)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при переборе строк: %w", err)
	}

	return DataCouples, nil
} // Получаем данные с таблиц
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
func deleteAllFromTable(db *sql.DB, tableName string) error {
	query := fmt.Sprintf("DELETE FROM %s;", tableName)

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("ошибка при удалении данных: %v", err)
	}

	fmt.Println("✅ ok - Данные удалены из таблицы", tableName)
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
func addRedactor(ctx context.Context, userID int64) bool {
	permCourse, permGroup := CourseGroupByUserID(userID)
	db, err := Connect(postgres)
	if err != nil {
		log.Println("addRedactor:", err)
	}
	defer db.Close()

	queryUpdate := `
		UPDATE users
		SET redactor = $1, permcourse = $2, permgroup = $3
		WHERE user_id = $4
	`
	_, err = db.ExecContext(ctx, queryUpdate, true, permCourse, permGroup, userID)
	if err != nil {
		fmt.Printf("ошибка при обновлении данных пользователя: %v", err)
		return false
	}

	fmt.Println("Данные успешно добавлены")
	return true
}
func deleteRedactor(ctx context.Context, userID int64) bool {
	db, err := Connect(postgres)
	if err != nil {
		log.Println("deleteRedactor:", err)
	}
	defer db.Close()

	queryUpdate := `
		UPDATE users
		SET redactor = $1
		WHERE user_id = $2
	`
	_, err = db.ExecContext(ctx, queryUpdate, false, userID)
	if err != nil {
		fmt.Printf("ошибка при обновлении данных пользователя: %v", err)
		return false
	}

	fmt.Println("Данные успешно добавлены")
	return true
}
func setRoleAdmin(ctx context.Context, userID int64, flag bool) bool {
	db, err := Connect(postgres)
	if err != nil {
		log.Println("setRoleAdmin:", err)
	}
	defer db.Close()

	queryUpdate := `
		UPDATE users
		SET admin = $1
		WHERE user_id = $2
	`
	_, err = db.ExecContext(ctx, queryUpdate, flag, userID)
	if err != nil {
		fmt.Printf("ошибка при обновлении данных пользователя: %v", err)
		return false
	}

	fmt.Println("Данные успешно добавлены")
	return true
}

func deleteAllSchedule() {
	allDataBases := FunctionDataBaseNames()
	for _, dbName := range allDataBases {
		db, err := Connect(dbName)
		if err != nil {
			log.Println("deleteAllSchedule:", err)
		}
		defer db.Close()

		allTables := FunctionTableNames(dbName)
		for _, table := range allTables {
			if table != "users" && table != "files" {
				deleteAllFromTable(db, table)
				fmt.Printf("Все данные из таблицы %s удалены\n", table)
			}
		}
	}
}
func schedule() {
	db, err := Connect(postgres)
	if err != nil {
		log.Println("schedule:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("Не удалось подключиться к серверу PostgreSQL: %v", err)
	}

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
		for _, groups := range groupsList {
			group := groups[0]
			group = strings.ToLower(replaceCyrillicWithLatin(removeSpaces(group)))
			connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", user, password, dbName, sslmode)
			courseDB, err := sql.Open("postgres", connStr)
			if err != nil {
				log.Printf("Не удалось подключиться к базе данных %s: %s\n", dbName, err)
			}
			defer db.Close()
			tableName := fmt.Sprintf(`"%s"`, group)

			createTableSQL := fmt.Sprintf(`
				CREATE TABLE %s (
					day_of_week TEXT,
					time TEXT,
					subject TEXT,
					auditory TEXT,
					teacher TEXT,
					weeks TEXT,
					added BOOLEAN DEFAULT FALSE,
					student TEXT DEFAULT '',
					userid BIGINT DEFAULT 0
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

					if couple != "" {
						key := fmt.Sprintf("%s|%s|%s", day, time, couple)

						if !contains(seen, key) {
							seen = append(seen, key)

							classCouple := parseClassInfo(couple)

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
func PinsByStarost(course, group, day, time, subject, auditory, teacher, weeks, student string, added bool, userID int64) string {
	db, err := Connect(renameSheetGroup(course))
	if err != nil {
		log.Println("PinsByStarost:", err)
		return ""
	}
	defer db.Close()
	var answer string
	tableName := fmt.Sprintf(`"%s"`, renameSheetGroup(group))

	_, err = db.Exec(fmt.Sprintf(`
	INSERT INTO %s (day_of_week, time, subject, auditory, teacher, weeks, added, student, userid)
	SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9
	WHERE NOT EXISTS (
		SELECT 1 FROM %s WHERE day_of_week = $1 AND time = $2 AND subject = $3 AND auditory = $4 AND teacher = $5 AND weeks = $6 AND added = $7 AND student = $8 AND userid = $9
		)
	`, tableName, tableName), day, time, subject, auditory, teacher, weeks, added, student, userID)
	if err != nil {
		answer = ""
		fmt.Printf("Ошибка при добавлении записи в таблицу для группы %s: %s\n", renameSheetGroup(group), err)
		return answer
	}
	answer = fmt.Sprintf("✅ Запись для группы %s успешно добавлена!\n", group)
	return answer
}
func getPermCourseGroupByUserID(userID int64) (string, string) {
	db, err := Connect(postgres)
	if err != nil {
		log.Println("getPermCourseGroupByUserID:", err)
		return "", ""
	}
	defer db.Close()

	query := `
    SELECT permcourse, permgroup
    FROM users
    WHERE user_id = $1 AND redactor = $2;
`
	var permCourse, permGroup string

	err = db.QueryRow(query, userID, true).Scan(&permCourse, &permGroup)
	if err == sql.ErrNoRows {
		fmt.Printf("Пользователь с ID %d не найден\n", userID)
		return "", ""
	} else if err != nil {
		fmt.Printf("Ошибка при выполнении запроса: %v\n", err)
		return "", ""
	}
	return permCourse, permGroup
}
func CourseGroupByUserID(userID int64) (string, string) {
	db, err := Connect(postgres)
	if err != nil {
		log.Println("CourseGroupByUserID:", err)
		return "", ""
	}
	defer db.Close()

	query := `
    SELECT user_course, user_group
    FROM users
    WHERE user_id = $1;
`
	var Course, Group string

	err = db.QueryRow(query, userID).Scan(&Course, &Group)
	if err == sql.ErrNoRows {
		fmt.Printf("Пользователь с ID %d не найден\n", userID)
		return "", ""
	} else if err != nil {
		fmt.Printf("Ошибка при выполнении запроса: %v\n", err)
		return "", ""
	}
	return Course, Group
}
func GetPinsByUserID(course, group string, userID int64) []string {

	if course == "" || group == "" {
		return []string{}
	}
	course = renameSheetGroup(course)
	group = renameSheetGroup(group)

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", user, password, course, sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Не удалось подключиться к базе данных %s: %s\n", course, err)
		return []string{}
	}
	defer db.Close()

	tableName := fmt.Sprintf(`"%s"`, group)

	query := fmt.Sprintf(`
		SELECT day_of_week, time, subject, auditory, teacher, weeks
		FROM %s
		WHERE userid = $1;
	`, tableName)

	rows, err := db.Query(query, userID)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	defer rows.Close()
	count := 0
	var listPins []string
	for rows.Next() {
		count += 1
		var day_of_week, time, subject, auditory, teacher, weeks string
		if err := rows.Scan(&day_of_week, &time, &subject, &auditory, &teacher, &weeks); err != nil {
			fmt.Println(err)
		}
		pin := fmt.Sprintf("<blockquote><b>[%d]</b>\n<b>%s</b><i> (%s)</i>\n<i>    📓 %s\n    🗝 %s\n    🪪 %s\n    🔍 %s</i></blockquote>\n", count, time, day_of_week, subject, auditory, teacher, weeks)
		listPins = append(listPins, pin)
	}

	if err = rows.Err(); err != nil {
		fmt.Println(err)
		return []string{}
	}
	return listPins
}
func GetRedactorsByUserID() []string {
	db, err := Connect(postgres)
	if err != nil {
		log.Println("GetRedactorsByUserID:", err)
		return []string{}
	}
	defer db.Close()

	query := `
		SELECT id, user_id, username, first_name, permcourse, permgroup
		FROM users
		WHERE redactor = $1
		ORDER BY id;
	`

	rows, err := db.Query(query, true)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	defer rows.Close()
	var listRedactors []string
	for rows.Next() {
		var id int
		var userid int64
		var username, firstname, permCourse, permGroup string
		if err := rows.Scan(&id, &userid, &username, &firstname, &permCourse, &permGroup); err != nil {
			fmt.Println(err)
		}
		redactor := fmt.Sprintf("<blockquote>%d.  %s\n     %s / %s\n     @%s (%d) </blockquote>\n", id, firstname, permCourse, permGroup, username, userid)
		listRedactors = append(listRedactors, redactor)
	}

	if err = rows.Err(); err != nil {
		fmt.Println(err)
		return []string{}
	}
	return listRedactors
}
func GetAdmins() []string {
	db, err := Connect(postgres)
	if err != nil {
		log.Println("GetAdmins:", err)
		return []string{}
	}
	defer db.Close()

	query := `
		SELECT id, user_id, username, first_name
		FROM users
		WHERE admin = $1
		ORDER BY id;
	`

	rows, err := db.Query(query, true)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	defer rows.Close()

	var listAdmins []string
	for rows.Next() {
		var id int
		var userid int64
		var username, firstname string
		if err := rows.Scan(&id, &userid, &username, &firstname); err != nil {
			fmt.Println(err)
		}
		admin := fmt.Sprintf("<blockquote>%d.  %s\n     @%s (%d) </blockquote>\n", id, firstname, username, userid)
		listAdmins = append(listAdmins, admin)
	}

	if err = rows.Err(); err != nil {
		fmt.Println(err)
		return []string{}
	}
	return listAdmins
}
func GetUsersAll() []string {
	user := ""
	db, err := Connect(postgres)
	if err != nil {
		log.Println("GetUsersAll:", err)
		return []string{}
	}
	defer db.Close()

	query := `
		SELECT id, user_id, username, first_name, redactor, user_course, user_group, admin
		FROM users
		ORDER BY id	
	`

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	defer rows.Close()

	var listUsers []string

	for rows.Next() {
		var (
			id         int
			userid     int64
			username   string
			firstname  string
			usercourse string
			usergroup  string
			redactor   bool
			admin      bool
		)
		if err := rows.Scan(&id, &userid, &username, &firstname, &redactor, &usercourse, &usergroup, &admin); err != nil {
			fmt.Println(err)
		}
		user = ""
		middle := ""
		usersStart := fmt.Sprintf("<blockquote>%d.  ", id)
		usersEnd := ""
		if redactor {
			middle += "✏️ "
		}
		if admin {
			middle += "🎫 "
		}
		if usercourse == "" || usergroup == "" {
			usersEnd = fmt.Sprintf("%s\n     @%s (%d) </blockquote>\n", firstname, username, userid)
		} else {
			usersEnd = fmt.Sprintf(" %s\n     @%s (%d)\n     %s / %s\n</blockquote>", firstname, username, userid, usercourse, usergroup)
		}
		user += usersStart + middle + usersEnd
		listUsers = append(listUsers, user)

	}
	if err = rows.Err(); err != nil {
		fmt.Println(err)
		return []string{}
	}
	return listUsers
}
func IsRedactorsByUserID(userID int64) bool {
	db, err := Connect(postgres)
	if err != nil {
		log.Println("IsRedactorsByUserID:", err)
		return false
	}
	defer db.Close()

	query := `
	SELECT redactor
	FROM users
	WHERE user_id = $1;
`
	var redactor bool

	err = db.QueryRow(query, userID).Scan(&redactor)
	if err == sql.ErrNoRows {
		fmt.Printf("Пользователь с ID %d не найден\n", userID)
		return false
	} else if err != nil {
		fmt.Printf("Ошибка при выполнении запроса: %v\n", err)
		return false
	}

	return redactor
}
func IsAdminByUserID(userID int64) bool {
	db, err := Connect(postgres)
	if err != nil {
		log.Println("IsAdminByUserID:", err)
		return false
	}
	defer db.Close()

	query := `
	SELECT admin
	FROM users
	WHERE user_id = $1;
`
	var admin bool

	err = db.QueryRow(query, userID).Scan(&admin)
	if err == sql.ErrNoRows {
		fmt.Printf("Пользователь с ID %d не найден\n", userID)
		return false
	} else if err != nil {
		fmt.Printf("Ошибка при выполнении запроса: %v\n", err)
		return false
	}

	return admin
}
func deleteNoticessByUserID(userID int64) error {
	course, group := getPermCourseGroupByUserID(userID)
	if course == "" || group == "" {
		return fmt.Errorf("error deleting records")
	}
	course = renameSheetGroup(course)
	group = renameSheetGroup(group)

	db, err := Connect(course)
	if err != nil {
		log.Println("deleteNoticessByUserID:", err)
		return err
	}
	defer db.Close()

	tableName := fmt.Sprintf(`"%s"`, group)

	query := fmt.Sprintf(`DELETE FROM %s WHERE userid = $1`, tableName)

	result, err := db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("error deleting records: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %w", err)
	}

	fmt.Printf("Deleted %d rows for userID: %d\n", rowsAffected, userID)
	return nil
}
func FunctionDataBaseNames() []string {
	db, err := Connect(postgres)
	if err != nil {
		log.Println("FunctionDataBaseNames:", err)
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

	db, err := Connect(postgres)
	if err != nil {
		log.Println("FunctionTableNames:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("TableNames: Не удалось подключиться к серверу PostgreSQL: %s", err)
	}

	fmt.Println("TableNames: Успешно подключено к серверу PostgreSQL!")

	return RequestGroupes(tbName)
} // Таблицы
func FunctionDataBaseReader() {
	db, err := Connect(postgres)
	if err != nil {
		log.Println("FunctionDataBaseReader:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("DataBaseReader: Не удалось подключиться к серверу PostgreSQL: %s", err)
	}

	fmt.Println("DataBaseReader: Успешно подключено к серверу PostgreSQL!")
}
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
	db, err := Connect(postgres)
	if err != nil {
		log.Println("getGroupsByCourseRu:", err)
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

	db, err := Connect(postgres)
	if err != nil {
		log.Println("FunctionDataBaseTableData:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("DTBC: Не удалось подключиться к серверу PostgreSQL: %s", err)
	}

	return RequestGroupDay(dbName, tbName, day)
} // Данные Расписания
func PinGroup(ctx context.Context, userID int64, course, group string) (string, error) {

	db, err := Connect(postgres)
	if err != nil {
		log.Println("PinGroup:", err)
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

	db, err := Connect(postgres)
	if err != nil {
		log.Println("GetUserCourseAndGroup:", err)
		return "", "", fmt.Errorf("%s %w", errorConnect, err)
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

	db, err := Connect(postgres)
	if err != nil {
		log.Println("getExcelName:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Printf("Не удалось подключиться к серверу PostgreSQL: %s", err)
	}

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
	db, err := Connect(postgres)
	if err != nil {
		log.Println("getAllSheets:", err)
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
	db, err := Connect(postgres)
	if err != nil {
		log.Println("ReloadFile:", err)
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
