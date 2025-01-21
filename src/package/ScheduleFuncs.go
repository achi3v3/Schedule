package funcExcel

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func FunctionScheduleFuncs() {
	fmt.Println("func: Schedule Funcs")
}

type ButtonAction struct {
	Label        string   `json:"label"`
	CallbackData string   `json:"callback_data"`
	NextButtons  []string `json:"next_buttons"`
}

// ===========================================BUTTONSHANDLERS==============================================================
func getAdjacentDays(currentDay string) (string, string) {
	daysOfWeek := []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}
	currentIndex := -1
	for i, day := range daysOfWeek {
		if day == currentDay {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return "", ""
	}
	prevIndex := (currentIndex - 1 + len(daysOfWeek)) % len(daysOfWeek)
	nextIndex := (currentIndex + 1) % len(daysOfWeek)

	return daysOfWeek[prevIndex], daysOfWeek[nextIndex]
} // Лево Право День недели
func createButtonActions(state map[string]string) map[string]ButtonAction {
	courses, _ := getAllSheets()
	// courses := get_sheets(get_file_excel())
	var buttonActions = make(map[string]ButtonAction)

	if state["course"] == "" {
		for _, course := range courses {
			if course != "Расписание" {
				groupsAl, _ := getGroupsByCourse(course)
				// groupsAl := get_groups(get_file_excel(), course)
				buttonActions[renameSheetGroup(course)] = ButtonAction{
					Label:        course,
					CallbackData: renameSheetGroup(course),
					NextButtons:  groupsAl,
				}
			}
		}
	}
	if state["course"] != "" && state["group"] == "" {
		groupAl, _ := getGroupsByCourse(state["course"])
		for _, group := range groupAl {
			buttonActions[renameSheetGroup(group)] = ButtonAction{
				Label:        group,
				CallbackData: renameSheetGroup(group),
				NextButtons:  get_days_for_couple(),
			}
		}
	}
	if state["course"] != "" && state["group"] != "" && state["day"] == "" {
		for _, day := range get_days_for_couple() {
			buttonActions[day] = ButtonAction{
				Label:        day,
				CallbackData: day,
				NextButtons:  []string{},
			}
		}
	}
	if state["course"] != "" && state["group"] != "" && state["day"] != "" {
		buttonActions["schedule"] = ButtonAction{
			Label:        "Schedule",
			CallbackData: "schedule",
			NextButtons:  []string{"Last"},
		}
	}

	return buttonActions
} // Реакции Кнопок
func dynamic_buttons(buttonLabels []string, state map[string]string) *tgbotapi.InlineKeyboardMarkup {
	var inlineButtons [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	for _, label := range buttonLabels {
		var button tgbotapi.InlineKeyboardButton
		if contains(get_days_for_couple(), label) {
			button = tgbotapi.NewInlineKeyboardButtonData(label, label)
		} else {
			button = tgbotapi.NewInlineKeyboardButtonData(label, label)
		}
		row = append(row, button)
		if len(row) == 3 || label == "Расписание" || label == "4 курс" || label == "Назад" {
			inlineButtons = append(inlineButtons, row)
			row = nil
		}
	}
	if len(row) > 0 {
		inlineButtons = append(inlineButtons, row)
	}
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: inlineButtons,
	}
} // Создание кнопок
func dynamic_buttonsFromActions(buttonActions map[string]ButtonAction, state map[string]string) *tgbotapi.InlineKeyboardMarkup {
	var buttonLabels []string
	var keys []string
	for key := range buttonActions {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		action := buttonActions[key]
		buttonLabels = append(buttonLabels, action.Label)
	}
	return dynamic_buttons(buttonLabels, state)
}
func handleButtonClick(update tgbotapi.Update, bot *tgbotapi.BotAPI, button string, state map[string]string) {
	// Проверяем, можно ли обработать запрос
	if !isRequestAllowed(int64(update.CallbackQuery.From.ID)) {
		// Отправляем сообщение, если пользователь превысил лимит запросов
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "🚫 Куда торопишься молодой.")
		bot.Send(msg)
		return
	}
	switch button {
	case "back_to_course":
		state["course"] = ""
		state["group"] = ""
		state["day"] = ""
	case "back_to_group":
		state["group"] = ""
		state["day"] = ""

	case "back_to_day":
		state["day"] = ""
	default:
		if contains(get_sheets(get_file_excel()), button) {
			state["course"] = button
		} else if isValidFormat(button) || isValidFormat(strings.Split(button, " ")[0]) {
			state["group"] = button
		} else if contains(get_days_for_couple(), button) {
			state["day"] = button
		}
	}
} // Обработка кнопок
// ===========================================DATABASEHANDLERS==============================================================
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

// ===========================================USERS==============================================================
func createTableUsers(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL UNIQUE,
		username TEXT NOT NULL,
		first_name TEXT NOT NULL,
		registration_time TIMESTAMP NOT NULL,
		user_group TEXT DEFAULT ''
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Ошибка при создании таблицы: %s", err)
	}
	fmt.Println("USERS OKAY")
}
func addUser(db *sql.DB, userID int, username, firstName string) bool {
	var exists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`
	err := db.QueryRow(queryCheck, userID).Scan(&exists)
	if err != nil {
		return false
	}
	if exists {
		return false
	}
	queryInsert := `
		INSERT INTO users (user_id, username, first_name, registration_time, user_group)
		VALUES ($1, $2, $3, $4, $5)
	`
	registrationTime := time.Now()
	userGroup := ""
	_, err = db.Exec(queryInsert, userID, username, firstName, registrationTime, userGroup)
	if err != nil {
		return false
	}
	fmt.Println("Пользователь успешно добавлен в базу данных!")
	return true
}

// ===========================================SHEETS==============================================================
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

// ===========================================GROUPS==============================================================
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
func getGroupsByCourse(courseName string) ([]string, error) {
	connStr := "user=postgres password=password sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Ошибка при подключении к баз %sе данных: ", err)
	}
	defer db.Close()

	var groups []string

	rows, err := db.Query("SELECT namegroup FROM groups WHERE sheet = $1", renameSheetGroup(courseName))
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

// ===========================================OPTIMIZATION==============================================================
var userRequests = make(map[int64]time.Time) // хранение времени последнего запроса пользователя
var requestCooldown = 1 * time.Second        // время задержки между запросами
func isRequestAllowed(userID int64) bool {
	now := time.Now()
	lastRequestTime, exists := userRequests[userID]
	if !exists {
		userRequests[userID] = now
		return true
	}
	if now.Sub(lastRequestTime) < requestCooldown {
		return false
	}
	userRequests[userID] = now
	return true
}

// ===========================================FILEPROCCESING==============================================================
func downloadFile(fileURL, savePath string) (string, error) {
	output, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer output.Close()

	resp, err := http.Get(fileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(output, resp.Body)
	if err != nil {
		return "", err
	}

	// Возвращаем имя файла
	return filepath.Base(savePath), nil
}
