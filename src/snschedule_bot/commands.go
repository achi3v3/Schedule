package functions

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	start           = "/start"
	snstartSchedule = "/snstart_schedule"
	uploadFile      = "/snupload_schedule"
	mygroup         = "/sn_mygroup"
	unPinGroup      = "/snunpin_group"

	idOwner = 5266257091
)

func isSnStartSchedule(update *models.Update) bool {
	return update.Message != nil && update.Message.Text == snstartSchedule
}
func snStartSchedule(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID

	if isSpamming(chatID) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "💢 Не торопись молодой..",
		})
		return
	}
	sendCourseSelectionWitoutEdit(ctx, b, chatID)

	userStates.Lock()
	if _, exists := userStates.data[chatID]; !exists {
		userStates.data[chatID] = make(map[string]string)
	}
	userStates.Unlock()

	setUserState(chatID, "course", "")
	setUserState(chatID, "group", "")
	setUserState(chatID, "day", "")

	userId := update.Message.Chat.ID
	username := update.Message.From.Username
	firstname := update.Message.From.FirstName
	flag := addUser(userId, username, firstname)
	if flag {
		if username == "" {
			username = "none"
		}
		if firstname == "" {
			firstname = "none"
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    idOwner,
			Text:      fmt.Sprintf("<b>✅ Новый пользователь</b>\n\n<blockquote>id: %d</blockquote>\nusername: @%s\n<blockquote>firstname: %s</blockquote>\n", userId, username, firstname),
			ParseMode: models.ParseModeHTML,
		})
	}
}
func isUploadFile(update *models.Update) bool {
	return update.Message != nil && update.Message.Text == uploadFile
}
func snUploadFile(ctx context.Context, b *bot.Bot, update *models.Update) {

	chatID := update.Message.Chat.ID
	if isSpamming(chatID) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "💢 Не торопись молодой..",
		})
		return
	}
	if chatID != idOwner {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "🔞 Недостаточно прав.",
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      "✳️ Загрузка расписания:\n<blockquote>Файл: <i>File.xlsx</i></blockquote>\n<blockquote>Неделя: <i>17</i></blockquote>",
		ParseMode: models.ParseModeHTML,
	})

	userStates.Lock()
	if _, exists := userStates.data[chatID]; !exists {
		userStates.data[chatID] = make(map[string]string)
	}
	userStates.Unlock()
}
func isMyGroup(update *models.Update) bool {
	return update.Message != nil && update.Message.Text == mygroup
}
func snMyGroup(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	if isSpamming(chatID) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "💢 Не торопись молодой..",
		})
		return
	}
	course, group, err := GetUserCourseAndGroup(ctx, chatID)
	if err != nil {
		fmt.Printf("PinGroup: %s\n", err)
	}
	if course != "" && group != "" {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      fmt.Sprintf("🪪 Ваша группа:\n<blockquote>Уровень обучения: %s\nГруппа: %s</blockquote>", course, group),
			ParseMode: models.ParseModeHTML,
		})
		setUserState(chatID, "course", course)
		setUserState(chatID, "group", group)

		state := getUserState(chatID)
		sendDaySelectionWithoutEdit(ctx, b, chatID, state)
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      "⭕️ За вами не закреплены курс и группа..",
			ParseMode: models.ParseModeHTML,
		})
		sendCourseSelectionWitoutEdit(ctx, b, chatID)
	}
}
func isStart(update *models.Update) bool {
	return update.Message != nil && update.Message.Text == start
}
func Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	if isSpamming(chatID) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "💢 Не торопись молодой..",
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		Text:      fmt.Sprintf("🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> (⚙️ Стадия-разработки)\n\nℹ️ Поиск расписания — %s\n\n👨‍👩‍👦‍👦 Моя группа — %s\n<i>1. Группа должна быть закреплена\n2. «Закрепить группу» можно при выборе группы нажав</i>\n\n🚮 Открепить группу — %s\n\n🆕 Загрузить расписание — %s\n<i>1. Ожидается эксель-файл + прикрепленный текст(неделя)\n2. Доступ только у администраторов</i>", snstartSchedule, mygroup, unPinGroup, uploadFile),
		ChatID:    chatID,
		ParseMode: models.ParseModeHTML,
	})
}
func isUnPin(update *models.Update) bool {
	return update.Message != nil && update.Message.Text == unPinGroup
}
func snUnPinGroup(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	if isSpamming(chatID) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "💢 Не торопись молодой..",
		})
		return
	}
	_, err := PinGroup(ctx, userID, "", "")
	if err != nil {
		fmt.Printf("UnPinGroup: %s\n", err)
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      "🪪 Ваша данные очищены:\n<blockquote>Уровень обучения: \nГруппа: </blockquote>",
		ParseMode: models.ParseModeHTML,
	})
}
