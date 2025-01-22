package funcExcel

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var userLastAction = struct {
	sync.RWMutex
	data map[int64]time.Time
}{
	data: make(map[int64]time.Time),
}

func getWeek() string {
	nameFileSlice, _ := getExcelName()
	GlobalWeek := nameFileSlice[1]
	return GlobalWeek
}

const actionCooldown = 400 * time.Millisecond

func isSpamming(userID int64) bool {
	userLastAction.Lock()
	defer userLastAction.Unlock()

	now := time.Now()
	lastAction, exists := userLastAction.data[userID]

	if exists && now.Sub(lastAction) < actionCooldown {
		return true
	}

	userLastAction.data[userID] = now
	return false
}
func sendCourseSelectionWitoutEdit(ctx context.Context, b *bot.Bot, chatID int64) {
	resetUserState(chatID)
	courses, _ := getAllSheets()

	var keyboardRows [][]models.InlineKeyboardButton
	row := []models.InlineKeyboardButton{}
	for i, course := range courses {
		if course == "Расписание" {
			continue
		}
		row = append(row, models.InlineKeyboardButton{Text: (course), CallbackData: course})
		if (i)%3 == 0 || course == "Расписание" || course == "4 курс" {
			keyboardRows = append(keyboardRows, row)
			row = []models.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		keyboardRows = append(keyboardRows, row)
	}

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}
	msgText := fmt.Sprintf("🏛 Расписание by <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> (⚙️ Бета-версия)\n📆 Установленная неделя: %s\n\nВыберите уровень обучения:", getWeek())

	sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        msgText,
		ReplyMarkup: keyboard,
		ParseMode:   models.ParseModeHTML,
	})
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
	setUserMessageID(chatID, sentMsg.ID)

}
func sendCourseSelection(ctx context.Context, b *bot.Bot, chatID int64) {
	resetUserState(chatID)
	courses, _ := getAllSheets()

	var keyboardRows [][]models.InlineKeyboardButton
	row := []models.InlineKeyboardButton{}
	for i, course := range courses {
		if course == "Расписание" {
			continue
		}
		row = append(row, models.InlineKeyboardButton{Text: (course), CallbackData: course})
		if (i)%3 == 0 || course == "Расписание" || course == "4 курс" {
			keyboardRows = append(keyboardRows, row)
			row = []models.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		keyboardRows = append(keyboardRows, row)
	}

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}
	msgText := fmt.Sprintf("🏛 Расписание by <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> (⚙️ Бета-версия)\n📆 Установленная неделя: %s\n\nВыберите уровень обучения:", getWeek())

	messageID, exists := getUserMessageID(chatID)
	if !exists {
		sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        msgText,
			ReplyMarkup: keyboard,
			ParseMode:   models.ParseModeHTML,
		})
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		setUserMessageID(chatID, sentMsg.ID)

	} else {
		editMessage(ctx, b, chatID, messageID, msgText, keyboard)
	}
}

func sendSchedule(ctx context.Context, b *bot.Bot, chatID int64, schedule string, state map[string]string) {

	var keyboardRows [][]models.InlineKeyboardButton

	prevday, nextday := getAdjacentDays(state["day"])

	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: prevday, CallbackData: prevday},
		{Text: "Назад", CallbackData: "back"},
		{Text: nextday, CallbackData: nextday},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}
	messageID, exists := getUserMessageID(chatID)
	if !exists {
		sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        schedule,
			ReplyMarkup: keyboard,
			ParseMode:   models.ParseModeHTML,
		})
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		setUserMessageID(chatID, sentMsg.ID)

	} else {
		editMessage(ctx, b, chatID, messageID, schedule, keyboard)
	}
}

func sendGroupSelection(ctx context.Context, b *bot.Bot, chatID int64, state map[string]string) {
	groups, _ := getGroupsByCourseRu(state["course"])
	var keyboardRows [][]models.InlineKeyboardButton
	row := []models.InlineKeyboardButton{}
	for i, group := range groups {
		row = append(row, models.InlineKeyboardButton{Text: group, CallbackData: group})

		if (i+1)%3 == 0 || i == len(group)-1 {
			keyboardRows = append(keyboardRows, row)
			row = []models.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		keyboardRows = append(keyboardRows, row)
	}
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "back"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	msgText := fmt.Sprintf("🏛 Расписание <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> (⚙️ Бета-версия)\n📆 Установленная неделя: %s\n\nУровень обучения: %s\nВыберите группу:", getWeek(), state["course"])

	messageID, exists := getUserMessageID(chatID)
	if !exists {
		sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        msgText,
			ReplyMarkup: keyboard,
			ParseMode:   models.ParseModeHTML,
		})
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		setUserMessageID(chatID, sentMsg.ID)

	} else {
		editMessage(ctx, b, chatID, messageID, msgText, keyboard)
	}
}

func sendDaySelection(ctx context.Context, b *bot.Bot, chatID int64, state map[string]string) {
	days := []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}
	var keyboardRows [][]models.InlineKeyboardButton
	row := []models.InlineKeyboardButton{}
	for i, day := range days {
		row = append(row, models.InlineKeyboardButton{Text: day, CallbackData: day})

		if (i+1)%3 == 0 || i == len(days)-1 {
			keyboardRows = append(keyboardRows, row)
			row = []models.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		keyboardRows = append(keyboardRows, row)
	}
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "back"},
	})

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	msgText := fmt.Sprintf("🏛 Расписание <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> (⚙️ Бета-версия)\n📆 Установленная неделя: %s\n\nУровень обучения: %s\nГруппа: %s\nВыберите день:", getWeek(), state["course"], state["group"])

	messageID, exists := getUserMessageID(chatID)
	if !exists {
		sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        msgText,
			ReplyMarkup: keyboard,
			ParseMode:   models.ParseModeHTML,
		})
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		setUserMessageID(chatID, sentMsg.ID)
	} else {
		editMessage(ctx, b, chatID, messageID, msgText, keyboard)
	}
}

func getSchedule(state map[string]string) string {

	allrangetime := []string{
		"08.00-09.20",
		"09.35-10.55",
		"11.35-12.55",
		"13.10-14.30",
		"15.10-16.30",
		"16.45-18.05",
		"18.20-19.40",
		"18.20-19.40",
		"19.55-21.15",
	}

	GlobalWeek := "17"
	startcoupleString := fmt.Sprintf("🏛 Расписание <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> (⚙️ Бета-версия)\n📆 Установленная неделя: %s\n\nУровень обучения: %s\nГруппа: %s\n\n<b>📅 %s</b>\n\n", GlobalWeek, state["course"], state["group"], state["day"])
	coupleList := FunctionDataBaseTableData(renameSheetGroup(state["course"]), renameSheetGroup(state["group"]), state["day"])
	coupleString, flagConcatenateAuditory, flagConcatenateTeacher := "", "", ""
	for i := 0; i < len(coupleList); i++ {
		numberCoupleTime := 0
		CoupleTime := coupleList[i][1]
		if contains(allrangetime, CoupleTime) {
			numberCoupleTime = findIndex(allrangetime, CoupleTime) + 1
		}
		CoupleSubject := coupleList[i][2]
		CoupleAuditory := coupleList[i][3]
		CoupleTeacher := coupleList[i][4]
		CoupleWeeks := coupleList[i][5]
		if flagConcatenateAuditory != "" {
			CoupleAuditory = fmt.Sprintf("%s / %s", CoupleAuditory, flagConcatenateAuditory)
			flagConcatenateAuditory = ""
		}
		if flagConcatenateTeacher != "" {
			CoupleTeacher = fmt.Sprintf("%s / %s", CoupleTeacher, flagConcatenateTeacher)
			flagConcatenateTeacher = ""
		}
		if i+1 < len(coupleList) {
			if CoupleTime == coupleList[i+1][1] && CoupleSubject == coupleList[i+1][2] {
				if CoupleTeacher == coupleList[i+1][4] && CoupleAuditory != coupleList[i+1][3] {
					flagConcatenateAuditory = CoupleAuditory
					continue
				} else if CoupleTeacher != coupleList[i+1][4] && CoupleAuditory != coupleList[i+1][3] {
					flagConcatenateAuditory = CoupleAuditory
					flagConcatenateTeacher = CoupleTeacher
					continue
				} else if CoupleTeacher != coupleList[i+1][4] && CoupleAuditory == coupleList[i+1][3] {
					flagConcatenateTeacher = CoupleTeacher
					continue
				}
			}
		}
		if CoupleWeeks != "—" {
			coupleString += fmt.Sprintf("<blockquote><b>%s</b> <i>(%d пара)\n</i>    📓 <i>%s</i>\n    🗝 <i>%s</i>\n    🪪 <i>%s</i>\n    🔍 <i>%s</i></blockquote>\n", CoupleTime, numberCoupleTime, CoupleSubject, removeBrackets(CoupleAuditory), CoupleTeacher, removeBrackets(CoupleWeeks))
		} else {
			coupleString += fmt.Sprintf("<blockquote><b>%s</b> <i>(%d пара)\n</i>    📓 <i>%s</i>\n    🗝 <i>%s</i>\n    🪪 <i>%s</i></blockquote>\n", CoupleTime, numberCoupleTime, CoupleSubject, removeBrackets(CoupleAuditory), CoupleTeacher)

		}
	}

	return startcoupleString + coupleString
}
