package funcExcel

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func UniversalHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	if update == nil {
		log.Println("Update is nil")
		return
	}
	if update.Message != nil {
		if update.Message.Document != nil {
			if update.Message.Chat.ID == idOwner {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "✅ Администратор",
				})
				handleDocument(ctx, b, update)

			} else {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "🚫 Недостаточно прав",
				})
			}
			return
		}
	}
	if update.CallbackQuery == nil {
		log.Println("CallbackQuery is nil")
		return
	}
	callbackQuery := update.CallbackQuery
	message := callbackQuery.Message.Message
	chatID := message.Chat.ID
	callbackData := callbackQuery.Data

	if isSpamming(chatID) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "💢 Не торопись молодой..",
		})
		return
	}

	state := getUserState(chatID)
	if state == nil {
		userStates.Lock()
		userStates.data[chatID] = make(map[string]string)
		userStates.Unlock()
		state = getUserState(chatID)
	}

	if callbackData == "back" {
		if state["day"] != "" && state["group"] != "" && state["course"] != "" {
			deleteUserState(chatID, "day")
			sendDaySelection(ctx, b, chatID, state)
		} else if state["group"] != "" && state["course"] != "" {
			deleteUserState(chatID, "day")
			deleteUserState(chatID, "group")
			sendGroupSelection(ctx, b, chatID, state)
		} else {
			deleteUserState(chatID, "day")
			deleteUserState(chatID, "group")
			deleteUserState(chatID, "course")
			sendCourseSelection(ctx, b, chatID)
		}
		return
	}
	if state["course"] == "" && state["group"] == "" && state["day"] == "" {
		setUserState(chatID, "course", callbackData)
		sendGroupSelection(ctx, b, chatID, state)
		return
	}
	if state["group"] == "" && state["day"] == "" {
		setUserState(chatID, "group", callbackData)
		sendDaySelection(ctx, b, chatID, state)
		return
	}
	if state["day"] == "" {
		setUserState(chatID, "day", callbackData)
		schedule := getSchedule(state)
		sendSchedule(ctx, b, chatID, schedule, state)
		return
	}
	if state["day"] != "" {
		setUserState(chatID, "day", callbackData)
		schedule := getSchedule(state)
		sendSchedule(ctx, b, chatID, schedule, state)
		return
	}
}

func databaseHandler() {
	// =============================================POSTGRESQL==============================================
	connStr := "user=postgres password=password sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Ошибка при подключении к базе данных: %s", err)
	}
	defer db.Close()
	createFilesTable(db)
	createTableUsers(db)
	createDataBasesExcel(db)
	// ==============================================FOLDER==============================================
	saveDir := "uploaded_files"
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		err := os.Mkdir(saveDir, 0755)
		if err != nil {
			log.Printf("Failed to create directory: %v", err)
		}
	}
}
