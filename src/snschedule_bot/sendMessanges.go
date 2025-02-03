package functions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	pageSize        = 8
	pageSizeNotices = 3
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

const actionCooldown = 800 * time.Millisecond

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
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "home"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}
	msgText := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n📆 Установленная неделя: %s\n\nВыберите уровень обучения:", getWeek())

	sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        msgText,
		ReplyMarkup: keyboard,
		ParseMode:   models.ParseModeHTML,
	})
	if err != nil {
		fmt.Printf("%s %v", errorSendMsg, err)
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
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "home"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}
	msg := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n📆 Установленная неделя: %s\n\nВыберите уровень обучения:", getWeek())
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendSchedule(ctx context.Context, b *bot.Bot, chatID int64, schedule string, state map[string]string) {

	var keyboardRows [][]models.InlineKeyboardButton

	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "✏️ Добавить", CallbackData: "Добавить"},
	})
	prevday, nextday := getAdjacentDays(state["day"])
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: prevday, CallbackData: prevday},
		{Text: "Назад", CallbackData: "back"},
		{Text: nextday, CallbackData: nextday},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}
	sendEditMessage(ctx, b, chatID, schedule, keyboard)
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

	msg := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n📆 Установленная неделя: %s\n\nУровень обучения: %s\nВыберите группу:", getWeek(), state["course"])
	sendEditMessage(ctx, b, chatID, msg, keyboard)
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
		{Text: "📌 Закрепить группу", CallbackData: "Закрепить группу"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "back"},
	})

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	msg := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n📆 Установленная неделя: %s\n\nУровень обучения: %s\nГруппа: %s\nВыберите день:", getWeek(), state["course"], state["group"])
	sendEditMessage(ctx, b, chatID, msg, keyboard)
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
		numberCoupleTime := "x"
		CoupleTime := coupleList[i][1]
		if contains(allrangetime, CoupleTime) {
			numberCoupleTime = fmt.Sprintf("%d", findIndex(allrangetime, CoupleTime)+1)
		}
		CoupleSubject := coupleList[i][2]
		CoupleAuditory := coupleList[i][3]
		CoupleTeacher := coupleList[i][4]
		CoupleWeeks := coupleList[i][5]
		flag_added := false
		if len(coupleList[i]) > 6 {
			flag_added = true
		}
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
		if !flag_added {
			if CoupleWeeks != "—" {
				coupleString += fmt.Sprintf("<blockquote><b>%s</b> <i>(%s пара)\n</i>    📓 <i>%s</i>\n    🗝 <i>%s</i>\n    🪪 <i>%s</i>\n    🔍 <i>%s</i></blockquote>\n", CoupleTime, numberCoupleTime, CoupleSubject, removeBrackets(CoupleAuditory), CoupleTeacher, removeBrackets(CoupleWeeks))
			} else {
				coupleString += fmt.Sprintf("<blockquote><b>%s</b> <i>(%s пара)\n</i>    📓 <i>%s</i>\n    🗝 <i>%s</i>\n    🪪 <i>%s</i></blockquote>\n", CoupleTime, numberCoupleTime, CoupleSubject, removeBrackets(CoupleAuditory), CoupleTeacher)
			}
		} else {
			CoupleStudent := coupleList[i][6]
			if CoupleWeeks != "—" {
				coupleString += fmt.Sprintf("<blockquote><b>%s</b> <i>(by %s)\n</i>    📓 <i>%s</i>\n    🗝 <i>%s</i>\n    🪪 <i>%s</i>\n    🔍 <i>%s</i></blockquote>\n", CoupleTime, CoupleStudent, CoupleSubject, removeBrackets(CoupleAuditory), CoupleTeacher, removeBrackets(CoupleWeeks))
			} else {
				coupleString += fmt.Sprintf("<blockquote><b>%s</b> <i>(by %s)\n</i>    📓 <i>%s</i>\n    🗝 <i>%s</i>\n    🪪 <i>%s</i></blockquote>\n", CoupleTime, CoupleStudent, CoupleSubject, removeBrackets(CoupleAuditory), CoupleTeacher)
			}
		}
	}

	return startcoupleString + coupleString
}
func NoticeSendDaySelection(ctx context.Context, b *bot.Bot, chatID int64, course, group string) {
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

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	msg := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n📆 Установленная неделя: %s\n\n<b>✏️ Добавления примечания</b>\nУровень обучения: %s\nГруппа: %s\nВыберите день:", getWeek(), course, group)
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendFile(b *bot.Bot, chatID int64, folderPath, fileName string) error {
	filePath := filepath.Join(folderPath, fileName)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл %s: %w", filePath, err)
	}
	defer file.Close()

	ctx := context.Background()

	_, err = b.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID: chatID,
		Document: &models.InputFileUpload{
			Filename: fileName,
			Data:     file,
		},
		Caption: "📂 Эксель-файл расписания",
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	fmt.Println("Файл успешно отправлен!")
	return nil
}
func sendStart(ctx context.Context, b *bot.Bot, chatID int64) {

	var keyboardRows [][]models.InlineKeyboardButton
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "🎓 Расписание", CallbackData: "Расписание"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "🔒 Моя группа", CallbackData: "Моя группа"},
		{Text: "🔓 Открепить группу", CallbackData: "Открепить группу"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "🪪 Панель редактора", CallbackData: "Уполномоченным"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "📂 Эксель-файл", CallbackData: "Отправить файл"},
		{Text: "📃 Информация", CallbackData: "Информация"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "⚙️ Панель управления", CallbackData: "Панель управления"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "🧾 Поддержка", URL: "https://t.me/sn_mira"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}
	msg := "🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n📜 Главное меню"
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendControPanel(ctx context.Context, b *bot.Bot, chatID int64) {
	var keyboardRows [][]models.InlineKeyboardButton
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Пользователи", CallbackData: "Пользователи"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Редакторы", CallbackData: "Редакторы"},
		{Text: "Администраторы", CallbackData: "Администраторы"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Загрузить расписание", CallbackData: "Загрузить расписание"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Информационный лист", CallbackData: "Информационный лист"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "home"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	msg := "🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n<b>📇 Панель управления</b>\n"
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendNotPermisions(ctx context.Context, b *bot.Bot, chatID int64) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "🚫 Недостаточно прав",
	})
}
func sendError(ctx context.Context, b *bot.Bot, chatID int64) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "❌ Ошибка!",
	})
}
func sendRedactorPanel(ctx context.Context, b *bot.Bot, chatID int64) {
	var keyboardRows [][]models.InlineKeyboardButton
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Мои записи", CallbackData: "Мои записи"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "home"},
	})

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	msg := "🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n📝 <b>Панель редактора</b>\n\n"
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendMyNotices(ctx context.Context, b *bot.Bot, chatID, userID int64, page int) {
	course, group := getPermCourseGroupByUserID(userID)
	notices := GetPinsByUserID(course, group, userID)

	var keyboardRows [][]models.InlineKeyboardButton
	noticesList := ""

	if len(notices) != 0 {
		totalNotices := len(notices)

		start := page * pageSizeNotices
		end := start + pageSizeNotices
		if end > totalNotices {
			end = totalNotices
		}
		if start >= totalNotices {
			return
		}
		if page > 0 || end < totalNotices {
			var navButtons []models.InlineKeyboardButton
			if page > 0 {
				navButtons = append(navButtons, models.InlineKeyboardButton{
					Text: "Предыдущий", CallbackData: "Мои записи:" + strconv.Itoa(page-1),
				})
			}
			if end < totalNotices {
				navButtons = append(navButtons, models.InlineKeyboardButton{
					Text: "Следующий", CallbackData: "Мои записи:" + strconv.Itoa(page+1),
				})
			}
			if len(navButtons) > 0 {
				keyboardRows = append(keyboardRows, navButtons)
			}
		}

		for _, notice := range notices[start:end] {
			noticesList += notice
		}

		keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
			{Text: "❌ Очистить записи", CallbackData: "Очистить записи"},
		})
	} else {

		keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
			{Text: "✏️ Добавить ", CallbackData: "Группа"},
		})
	}
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "Уполномоченным"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	msg := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\nУровень обучения: %s\nГруппа: %s\n\n📝 <b>Ваши записи (%d)</b>\n\n<i>Лист %d:</i>\n%s", course, group, userID, page+1, noticesList)
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendNoticesByUserID(ctx context.Context, b *bot.Bot, chatID, userID int64, page int) {
	course, group := getPermCourseGroupByUserID(userID)
	notices := GetPinsByUserID(course, group, userID)

	var keyboardRows [][]models.InlineKeyboardButton
	noticesList := ""

	if len(notices) != 0 {
		totalNotices := len(notices)

		start := page * pageSizeNotices
		end := start + pageSizeNotices
		if end > totalNotices {
			end = totalNotices
		}
		if start >= totalNotices {
			return
		}

		if page > 0 || end < totalNotices {
			var navButtons []models.InlineKeyboardButton
			if page > 0 {
				navButtons = append(navButtons, models.InlineKeyboardButton{
					Text: "Предыдущий", CallbackData: "Просмотреть записи:" + strconv.Itoa(page-1),
				})
			}
			if end < totalNotices {
				navButtons = append(navButtons, models.InlineKeyboardButton{
					Text: "Следующий", CallbackData: "Просмотреть записи:" + strconv.Itoa(page+1),
				})
			}
			if len(navButtons) > 0 {
				keyboardRows = append(keyboardRows, navButtons)
			}
		}
		for _, notice := range notices[start:end] {
			noticesList += notice
		}

		keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
			{Text: "❌ Очистить записи", CallbackData: "Очистить записи пользователя"},
		})
	}
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "Просмотр пользователя"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	msg := fmt.Sprintf("<b>🔰 Управление пользователем</b> %d\n\nУровень обучения: %s\nГруппа: %s\n\n📝 <b>Записи пользователя</b>\n\n<i>Лист %d:</i>\n%s", userID, course, group, page+1, noticesList)
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendGetRedactors(ctx context.Context, b *bot.Bot, chatID int64, page int) {
	redactors := GetRedactorsByUserID()
	totalUsers := len(redactors)

	start := page * pageSize
	end := start + pageSize
	if end > totalUsers {
		end = totalUsers
	}
	if start >= totalUsers {
		return
	}
	var keyboardRows [][]models.InlineKeyboardButton

	if page > 0 || end < totalUsers {
		var navButtons []models.InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, models.InlineKeyboardButton{
				Text: "Предыдущий", CallbackData: "Редакторы:" + strconv.Itoa(page-1),
			})
		}
		if end < totalUsers {
			navButtons = append(navButtons, models.InlineKeyboardButton{
				Text: "Следующий", CallbackData: "Редакторы:" + strconv.Itoa(page+1),
			})
		}
		if len(navButtons) > 0 {
			keyboardRows = append(keyboardRows, navButtons)
		}
	}
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "Панель управления"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	userList := ""
	for _, user := range redactors[start:end] {
		userList += user
	}

	msg := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n📝 <b>Редакторы</b>\n\n<i>Лист %d:</i>\n%s", page+1, userList)
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendGetAdmins(ctx context.Context, b *bot.Bot, chatID int64, page int) {
	admins := GetAdmins()
	totalUsers := len(admins)

	start := page * pageSize
	end := start + pageSize
	if end > totalUsers {
		end = totalUsers
	}
	if start >= totalUsers {
		return
	}
	var keyboardRows [][]models.InlineKeyboardButton

	if page > 0 || end < totalUsers {
		var navButtons []models.InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, models.InlineKeyboardButton{
				Text: "Предыдущий", CallbackData: "Администраторы:" + strconv.Itoa(page-1),
			})
		}
		if end < totalUsers {
			navButtons = append(navButtons, models.InlineKeyboardButton{
				Text: "Следующий", CallbackData: "Администраторы:" + strconv.Itoa(page+1),
			})
		}
		if len(navButtons) > 0 {
			keyboardRows = append(keyboardRows, navButtons)
		}
	}
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "Панель управления"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	userList := ""
	for _, user := range admins[start:end] {
		userList += user
	}
	msg := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n🎫  <b>Администраторы</b>\n\n<i>Лист %d:</i>\n%s", page+1, userList)
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendUsers(ctx context.Context, b *bot.Bot, chatID int64, page int) {
	users := GetUsersAll()
	totalUsers := len(users)

	start := page * pageSize
	end := start + pageSize
	if end > totalUsers {
		end = totalUsers
	}
	if start >= totalUsers {
		return
	}
	var keyboardRows [][]models.InlineKeyboardButton

	if page > 0 || end < totalUsers {
		var navButtons []models.InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, models.InlineKeyboardButton{
				Text: "Предыдущий", CallbackData: "Пользователи:" + strconv.Itoa(page-1),
			})
		}
		if end < totalUsers {
			navButtons = append(navButtons, models.InlineKeyboardButton{
				Text: "Следующий", CallbackData: "Пользователи:" + strconv.Itoa(page+1),
			})
		}
		if len(navButtons) > 0 {
			keyboardRows = append(keyboardRows, navButtons)
		}
	}
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "Панель управления"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	userList := ""
	for _, user := range users[start:end] {
		userList += user
	}
	msg := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n<i>✏️ — Редактор\n🎫 — Администратор</i>\n\n📝 <b>Список пользователей</b>\n\n<i>Лист %d:</i>\n%s", page+1, userList)
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendUploadFile(ctx context.Context, b *bot.Bot, chatID int64) {
	var keyboardRows [][]models.InlineKeyboardButton

	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "Панель управления"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	msg := "🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n✳️ Загрузка расписания\n<blockquote>Файл: <i>File.xlsx</i></blockquote>\n<blockquote>Неделя: <i>17</i></blockquote>"
	sendEditMessage(ctx, b, chatID, msg, keyboard)

}
func sendInfo(ctx context.Context, b *bot.Bot, chatID int64) {
	var keyboardRows [][]models.InlineKeyboardButton

	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "home"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	msg := "🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n📃 <b>Информационный лист</b>\n\n<blockquote><i>    Бот для поиска расписания, основан на считывании информации с Excel-файла.</i>\n\n    <b>🪪 Панель редактора</b> — <u>Панель доступна пользователям, которым назначили доступ для добавления записей в расписание</u> <i>(если в вашей группе нет такого пользователя и вы хотите стать им или назначить человека, напишите в поддержку)</i>.\n\n    <b>📂 Эксель-файл</b> — Отправляется эксель-файл, с которого были считаны неделя и данное расписание.\n\n    <b>⚙️ Панель управления</b> — <u>Панель доступна пользователям, с правами администратора</u>.\n\n    <b>🧾 Поддержка</b> — В поддержку можно обратиться по любым вопросам, относящимся к данному боту.</blockquote>"
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendUpdatePermisions(ctx context.Context, b *bot.Bot, chatID, userID int64) {
	course, group := CourseGroupByUserID(userID)
	var (
		keyboardRows [][]models.InlineKeyboardButton
		roles        string
	)
	if !IsRedactorsByUserID(userID) {
		keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
			{Text: "Сделать редактором", CallbackData: "Сделать редактором"},
		})
	} else {
		roles += "Редактор"
		keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
			{Text: "Убрать редактора", CallbackData: "Убрать редактора"},
		})
	}
	if !IsAdminByUserID(userID) {
		keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
			{Text: "Назначить админом", CallbackData: "Сделать админом"},
		})
	} else {
		if roles != "" {
			roles += " / Администратор"
		} else {
			roles += "Администратор"
		}
		keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
			{Text: "Убрать админку", CallbackData: "Убрать админку"},
		})
	}
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Просмотреть записи", CallbackData: "Просмотреть записи"},
	})
	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "Панель управления"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}
	msg := fmt.Sprintf("<b>🔰 Управление пользователем</b> %d\n\nУровень обучения: %s\nГруппа: %s\nРоли: %s\n", userID, course, group, roles)
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendRequestForSetRoleAdmin(ctx context.Context, b *bot.Bot, chatID, WhoID, userID int64) {
	msgUser := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n<b>Уведомление</b>\n\n<blockquote><i>Вы предложили назначение роли «Администратор» пользователю %d</i></blockquote>", userID)
	sendOnlyMessage(ctx, b, chatID, msgUser)
	msgOwner := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n<b>Уведомление</b>\n\n<blockquote>Администратор %d просит о назначении роли «Администратор» для %d</blockquote>", WhoID, userID)
	sendOnlyMessage(ctx, b, idOwner, msgOwner)
}
func sendAddRole(ctx context.Context, b *bot.Bot, userID int64, role string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    userID,
		Text:      fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n<b>Уведомление</b>\n\n<blockquote>Вам назначили роль «<i>%s</i>»</blockquote>", role),
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		msg := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n<b>Уведомление</b>\n\n<blockquote>До пользователя %d не дошло сообщение о назначении роли администратора</blockquote>", userID)
		sendOnlyMessage(ctx, b, idOwner, msg)
	}
}
func sendDeleteRole(ctx context.Context, b *bot.Bot, userID int64, role string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    userID,
		Text:      fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n<b>Уведомление</b>\n\n<blockquote>У вас отобрали роль «<i>%s</i>»</blockquote>", role),
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		msg := fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n<b>Уведомление</b>\n\n<blockquote>До пользователя %d не дошло сообщение о назначении роли администратора</blockquote>", userID)
		sendOnlyMessage(ctx, b, idOwner, msg)
	}
}
func sendAdminInfo(ctx context.Context, b *bot.Bot, chatID int64) {
	var keyboardRows [][]models.InlineKeyboardButton

	keyboardRows = append(keyboardRows, []models.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "Панель управления"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}
	msg := "🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)\n\n📃 <b>Администраторский информационный лист</b>\n\n<blockquote>ℹ️ Для управления пользователем введите в бота id-пользователя <i>(id можно взять в скобках в списке пользователей)</i>\n\nℹ️ «Сделать редактором» — у пользователя должна быть прикреплена группа, в которою он впоследствии сможет вносить записи.\n\nℹ️ «Назначить администратором» — отправляется запрос о назначении роли «Администратор».\n\nℹ️ «Убрать админку» — может только основной администратор\n\nℹ️ «Загрузить расписание» — может только уполномоченный пользователь.</blockquote>"
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}

func sendEditMessage(ctx context.Context, b *bot.Bot, chatID int64, msg string, keyboard *models.InlineKeyboardMarkup) {
	messageID, exists := getUserMessageID(chatID)

	if !exists {
		sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        msg,
			ReplyMarkup: keyboard,
			ParseMode:   models.ParseModeHTML,
		})
		if err != nil {
			fmt.Printf("%s %v", errorSendMsg, err)
		}
		setUserMessageID(chatID, sentMsg.ID)

	} else {
		err := editMessage(ctx, b, chatID, messageID, msg, keyboard)
		if err != nil {
			sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      chatID,
				Text:        msg,
				ReplyMarkup: keyboard,
				ParseMode:   models.ParseModeHTML,
			})
			if err != nil {
				fmt.Printf("Ошибка отправки сообщения: %v", err)

			}
			setUserMessageID(chatID, sentMsg.ID)
		}
	}
}
func sendOnlyMessage(ctx context.Context, b *bot.Bot, chatID int64, msg string) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
	})
}

func sendNotifyRole(ctx context.Context, b *bot.Bot, userID, setUserID int64, symbol, Text string) {
	msg := fmt.Sprintf("🔔 %d:\n<blockquote>%d %s %s</blockquote>", userID, setUserID, symbol, Text)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    idOwner,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
	})
}
