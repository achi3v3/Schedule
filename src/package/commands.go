package funcExcel

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	start           = "start"
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

}

func isUploadFile(update *models.Update) bool {
	return update.Message != nil && update.Message.Text == uploadFile
}

// func snUploadSchedule(ctx context.Context, b *bot.Bot, update *models.Update) {

// 	chatID := update.Message.Chat.ID
// 	if chatID != idOwner {
// 		sendMsgError(ctx, b, update)
// 	}
// 	sendCourseSelection(ctx, b, chatID)

// 	userStates.Lock()
// 	if _, exists := userStates.data[chatID]; !exists {
// 		userStates.data[chatID] = make(map[string]string)
// 	}
// 	userStates.Unlock()

// }

func isStart(update *models.Update) bool {
	return update.Message != nil && update.Message.Text == start
}
