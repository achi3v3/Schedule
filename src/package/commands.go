package funcExcel

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

	idOwner = 5266257091
)

func isSnStartSchedule(update *models.Update) bool {
	return update.Message != nil && update.Message.Text == snstartSchedule
}
func snStartSchedule(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
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

func isStart(update *models.Update) bool {
	return update.Message != nil && update.Message.Text == start
}
func Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      "🏛 <a href=\"https://t.me/sn_schedulebot\">Schedule Bot</a> (⚙️ Стадия-разработки)\n\nПоиск расписания — /snstart_schedule ",
		ParseMode: models.ParseModeHTML,
	})
}
