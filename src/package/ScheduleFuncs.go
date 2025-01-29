package funcExcel

import (
	"fmt"
	"io"
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

// ===========================================USERS==============================================================

// ===========================================SHEETS==============================================================

// ===========================================GROUPS==============================================================

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
