package functions

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	// Common message parts
	botHeader = "🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> — Расписание (⚙️ Бета-версия)"
	weekInfo  = "📆 Установленная неделя: %s"

	// Course and group info
	courseInfo = "\n\nУровень обучения: %s"
	groupInfo  = "\nГруппа: %s"

	// Page info
	pageInfo = "\n\n<i>Лист %d:</i>"

	// Common buttons
	backButton = "Назад"

	// Page sizes
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
func sendCourseSelection(ctx context.Context, b *bot.Bot, chatID int64, edit bool) {
	resetUserState(chatID)
	msg := fmt.Sprintf("%s\n%s\n\nВыберите уровень обучения:", botHeader, fmt.Sprintf(weekInfo, getWeek()))

	if edit {
		sendEditMessage(ctx, b, chatID, msg, CourseSelection())
	} else {
		sentMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        msg,
			ReplyMarkup: CourseSelection(),
			ParseMode:   models.ParseModeHTML,
		})
		if err != nil {
			fmt.Printf("%s %v", errorSendMsg, err)
		}
		setUserMessageID(chatID, sentMsg.ID)
	}
}

func sendSchedule(ctx context.Context, b *bot.Bot, chatID int64, schedule string, state map[string]string) {
	prevday, nextday := getAdjacentDays(state["day"])
	sendEditMessage(ctx, b, chatID, schedule, ScheduleKeyboard(prevday, nextday))
}
func sendGroupSelection(ctx context.Context, b *bot.Bot, chatID int64, state map[string]string) {
	msg := fmt.Sprintf("%s\n%s\n\nУровень обучения: %s\nВыберите группу:",
		botHeader,
		fmt.Sprintf(weekInfo, getWeek()),
		state["course"])
	sendEditMessage(ctx, b, chatID, msg, GroupSelection(state["course"]))
}
func sendDaySelection(ctx context.Context, b *bot.Bot, chatID int64, state map[string]string) {
	msg := fmt.Sprintf("%s\n%s\n\nУровень обучения: %s\nГруппа: %s\nВыберите день:",
		botHeader,
		fmt.Sprintf(weekInfo, getWeek()),
		state["course"],
		state["group"])
	sendEditMessage(ctx, b, chatID, msg, DaySelection())
}

type Schedule struct {
	Time    string
	Subject string
	Teacher string
	Room    string
	Weeks   string
	Student string
}

func convertToSchedule(couple []string) Schedule {
	schedule := Schedule{
		Time:    couple[1],
		Subject: couple[2],
		Room:    couple[3],
		Teacher: couple[4],
		Weeks:   couple[5],
	}
	if len(couple) > 6 {
		schedule.Student = couple[6]
	}
	return schedule
}
func (s Schedule) getPairNumber() string {
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

	if idx := findIndex(allrangetime, s.Time); idx != -1 {
		return fmt.Sprintf("%d", idx+1)
	}
	return "x"
}
func (s Schedule) HasStudent() bool {
	return s.Student != ""
}

func (s Schedule) FormatMessage() string {
	if s.HasStudent() {
		return fmt.Sprintf("<blockquote><b>%s</b> <i>(by %s)\n</i>    📓 <i>%s</i>\n    🗝 <i>%s</i>\n    🪪 <i>%s</i></blockquote>\n",
			s.Time, s.Student, s.Subject, s.Room, s.Teacher)
	}
	return fmt.Sprintf("<blockquote><b>%s</b> <i>(%s пара)\n</i>    📓 <i>%s</i>\n    🗝 <i>%s</i>\n    🪪 <i>%s</i></blockquote>\n",
		s.Time, s.getPairNumber(), s.Subject, s.Room, s.Teacher)
}

func getSchedule(state map[string]string) string {
	startcoupleString := fmt.Sprintf("%s\n%s\n\nУровень обучения: %s\nГруппа: %s\n\n<b>%s</b>\n\n",
		botHeader,
		fmt.Sprintf(weekInfo, getWeek()),
		state["course"],
		state["group"],
		state["day"])

	coupleList := FunctionDataBaseTableData(renameSheetGroup(state["course"]), renameSheetGroup(state["group"]), state["day"])
	var schedules []Schedule

	// Convert raw data to Schedule structs and handle concatenation
	for i := 0; i < len(coupleList); i++ {
		schedule := convertToSchedule(coupleList[i])

		// Check if we need to concatenate with next entry
		if i+1 < len(coupleList) {
			nextSchedule := convertToSchedule(coupleList[i+1])
			if schedule.Time == nextSchedule.Time && schedule.Subject == nextSchedule.Subject {
				if schedule.Teacher == nextSchedule.Teacher && schedule.Room != nextSchedule.Room {
					schedule.Room = fmt.Sprintf("%s / %s", schedule.Room, nextSchedule.Room)
					i++ // Skip next entry
				} else if schedule.Teacher != nextSchedule.Teacher && schedule.Room != nextSchedule.Room {
					schedule.Room = fmt.Sprintf("%s / %s", schedule.Room, nextSchedule.Room)
					schedule.Teacher = fmt.Sprintf("%s / %s", schedule.Teacher, nextSchedule.Teacher)
					i++ // Skip next entry
				} else if schedule.Teacher != nextSchedule.Teacher && schedule.Room == nextSchedule.Room {
					schedule.Teacher = fmt.Sprintf("%s / %s", schedule.Teacher, nextSchedule.Teacher)
					i++ // Skip next entry
				}
			}
		}

		schedules = append(schedules, schedule)
	}

	// Format all schedules
	var coupleString string
	for _, schedule := range schedules {
		coupleString += schedule.FormatMessage()
	}

	return startcoupleString + coupleString
}
func NoticeSendDaySelection(ctx context.Context, b *bot.Bot, chatID int64, course, group string) {

	var keyboardRows [][]models.InlineKeyboardButton
	days := []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}
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

	msg := fmt.Sprintf("%s\n%s\n%s%s%s\n<b>✏️ Добавления примечания</b>\nВыберите день:", botHeader, getWeek(), courseInfo, groupInfo, weekInfo)

	sendEditMessage(ctx, b, chatID, msg, keyboard)
}

func sendFile(b *bot.Bot, chatID int64, folderPath, fileName string) error {
	filePath := filepath.Join(folderPath, fileName)

	file, err := os.Open(filePath)
	if err != nil {
		logError(err, "sendFile: открытие файла")
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
		logError(err, "sendFile: отправка документа")
		return fmt.Errorf("%w", err)
	}

	return nil
}

func sendStart(ctx context.Context, b *bot.Bot, chatID int64) {
	msg := fmt.Sprintf("%s\n\n📜 Главное меню", botHeader)
	sendEditMessage(ctx, b, chatID, msg, StartKeyboard())
}
func sendControPanel(ctx context.Context, b *bot.Bot, chatID int64) {
	msg := fmt.Sprintf("%s\n\n<b>📇 Панель управления</b>\n", botHeader)
	sendEditMessage(ctx, b, chatID, msg, ControlPanelKeyboard())
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
	msg := fmt.Sprintf("%s\n\n📝 <b>Панель редактора</b>\n\n", botHeader)
	sendEditMessage(ctx, b, chatID, msg, RedactorPanelKeyboard())
}
func sendMyNotices(ctx context.Context, b *bot.Bot, chatID, userID int64, page int) {
	course, group := getPermCourseGroupByUserID(userID)
	notices := GetPinsByUserID(course, group, userID)
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

		for _, notice := range notices[start:end] {
			noticesList += notice
		}
	}

	msg := fmt.Sprintf("%s\n\nУровень обучения: %s\nГруппа: %s\n\n📝 <b>Ваши записи (%d)</b>%s\n%s",
		botHeader,
		course,
		group,
		userID,
		fmt.Sprintf(pageInfo, page+1),
		noticesList)
	sendEditMessage(ctx, b, chatID, msg, NoticesKeyboard(page, len(notices), true))
}
func sendNoticesByUserID(ctx context.Context, b *bot.Bot, chatID, userID int64, page int) {

	course, group := getPermCourseGroupByUserID(userID)
	notices := GetPinsByUserID(course, group, userID)
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

		for _, notice := range notices[start:end] {
			noticesList += notice
		}
	}

	msg := fmt.Sprintf("<b>🔰 Управление пользователем</b> %d\n\nУровень обучения: %s\nГруппа: %s\n\n📝 <b>Записи пользователя</b>%s\n%s",
		userID,
		course,
		group,
		fmt.Sprintf(pageInfo, page+1),
		noticesList)
	sendEditMessage(ctx, b, chatID, msg, NoticesKeyboard(page, len(notices), false))
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
		{Text: backButton, CallbackData: "Панель управления"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	userList := ""
	for _, user := range redactors[start:end] {
		userList += user
	}

	msg := fmt.Sprintf("%s\n\n📝 <b>Редакторы</b>\n%s", botHeader, userList)
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
		{Text: backButton, CallbackData: "Панель управления"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	userList := ""
	for _, user := range admins[start:end] {
		userList += user
	}
	msg := fmt.Sprintf("%s\n\n🎫  <b>Администраторы</b>\n%s", botHeader, userList)
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
		{Text: backButton, CallbackData: "Панель управления"},
	})
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}

	userList := ""
	for _, user := range users[start:end] {
		userList += user
	}
	msg := fmt.Sprintf("%s\n\n<i>✏️ — Редактор\n🎫 — Администратор</i>\n\n📝 <b>Список пользователей</b>\n%s", botHeader, userList)
	sendEditMessage(ctx, b, chatID, msg, keyboard)
}
func sendUploadFile(ctx context.Context, b *bot.Bot, chatID int64) {
	msg := fmt.Sprintf("%s\n\n✳️ Загрузка расписания\n<blockquote>Файл: <i>File.xlsx</i></blockquote>\n<blockquote>Неделя: <i>17</i></blockquote>", botHeader)
	sendEditMessage(ctx, b, chatID, msg, BackKeyboard("Панель управления"))
}
func sendInfo(ctx context.Context, b *bot.Bot, chatID int64) {
	msg := fmt.Sprintf("%s\n\n📃 <b>Информационный лист</b>\n\n<blockquote><i>    Бот для поиска расписания, основан на считывании информации с Excel-файла.</i>\n\n    <b>🪪 Панель редактора</b> — <u>Панель доступна пользователям, которым назначили доступ для добавления записей в расписание</u> <i>(если в вашей группе нет такого пользователя и вы хотите стать им или назначить человека, напишите в поддержку)</i>.\n\n    <b>📂 Эксель-файл</b> — Отправляется эксель-файл, с которого были считаны неделя и данное расписание.\n\n    <b>⚙️ Панель управления</b> — <u>Панель доступна пользователям, с правами администратора</u>.\n\n    <b>🧾 Поддержка</b> — В поддержку можно обратиться по любым вопросам, относящимся к данному боту.</blockquote>", botHeader)
	sendEditMessage(ctx, b, chatID, msg, BackKeyboard("home"))
}
func sendUpdatePermisions(ctx context.Context, b *bot.Bot, chatID, userID int64) {
	course, group := CourseGroupByUserID(userID)
	var roles string

	if IsRedactorsByUserID(userID) {
		roles += "Редактор"
	}
	if IsAdminByUserID(userID) {
		if roles != "" {
			roles += " / Администратор"
		} else {
			roles += "Администратор"
		}
	}

	msg := fmt.Sprintf("<b>🔰 Управление пользователем</b> %d\n\nУровень обучения: %s\nГруппа: %s\nРоли: %s\n",
		userID,
		course,
		group,
		roles)
	sendEditMessage(ctx, b, chatID, msg, UserPermissionsKeyboard(IsRedactorsByUserID(userID), IsAdminByUserID(userID)))
}
func sendRequestForSetRoleAdmin(ctx context.Context, b *bot.Bot, chatID, WhoID, userID int64) {
	msgUser := fmt.Sprintf("%s\n\n<b>Уведомление</b>\n\n<blockquote><i>Вы предложили назначение роли «Администратор» пользователю %d</i></blockquote>", botHeader, userID)
	sendOnlyMessage(ctx, b, chatID, msgUser)
	msgOwner := fmt.Sprintf("%s\n\n<b>Уведомление</b>\n\n<blockquote>Администратор %d просит о назначении роли «Администратор» для %d</blockquote>", botHeader, WhoID, userID)
	sendOnlyMessage(ctx, b, idOwner, msgOwner)
}

func sendAddRole(ctx context.Context, b *bot.Bot, userID int64, role string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    userID,
		Text:      fmt.Sprintf("%s\n\n<b>Уведомление</b>\n\n<blockquote>Вам назначили роль «<i>%s</i>»</blockquote>", botHeader, role),
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		logError(err, "sendAddRole: отправка уведомления")
		msg := fmt.Sprintf("\n\n<b>Уведомление</b>\n\n<blockquote>До пользователя %d не дошло сообщение о назначении роли администратора</blockquote>", userID)
		sendOnlyMessage(ctx, b, idOwner, msg)
	}
}

func sendDeleteRole(ctx context.Context, b *bot.Bot, userID int64, role string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    userID,
		Text:      fmt.Sprintf("%s\n\n<b>Уведомление</b>\n\n<blockquote>У вас отобрали роль «<i>%s</i>»</blockquote>", botHeader, role),
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		logError(err, "sendDeleteRole: отправка уведомления")
		msg := fmt.Sprintf("\n\n<b>Уведомление</b>\n\n<blockquote>До пользователя %d не дошло сообщение о назначении роли администратора</blockquote>", userID)
		sendOnlyMessage(ctx, b, idOwner, msg)
	}
}

func sendAdminInfo(ctx context.Context, b *bot.Bot, chatID int64) {
	msg := fmt.Sprintf("%s\n\n📃 <b>Администраторский информационный лист</b>\n\n<blockquote>ℹ️ Для управления пользователем введите в бота id-пользователя <i>(id можно взять в скобках в списке пользователей)</i>\n\nℹ️ «Сделать редактором» — у пользователя должна быть прикреплена группа, в которою он впоследствии сможет вносить записи.\n\nℹ️ «Назначить администратором» — отправляется запрос о назначении роли «Администратор».\n\nℹ️ «Убрать админку» — может только основной администратор\n\nℹ️ «Загрузить расписание» — может только уполномоченный пользователь.</blockquote>", botHeader)
	sendEditMessage(ctx, b, chatID, msg, BackKeyboard("Панель управления"))
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
			logError(err, "sendEditMessage: отправка сообщения")
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
				logError(err, "sendEditMessage: повторная отправка сообщения")
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

// Простая функция для логирования ошибок
func logError(err error, context string) {
	if err != nil {
		log.Printf("[ERROR] %s: %v", context, err)
	}
}
