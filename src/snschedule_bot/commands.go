package functions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	start           = "/start"
	snstartSchedule = "/snstart_schedule"
	uploadFile      = "/snupload_schedule"
	mygroup         = "/sn_mygroup"
	unPinGroup      = "/snunpin_group"
	getNotices      = "/snget_mynotices"
	getRedactors    = "/snget_redactors"
	idOwner         = 5266257091
)

func isStart(update *models.Update) bool {
	return update.Message != nil && update.Message.Text == start
}
func Start(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	if update.Message.Chat.Type != "private" {
		return // Игнорируем сообщения из групп
	}
	chatID := update.Message.Chat.ID
	deleteUserMessageID(chatID)

	if !checkUserActivity(chatID) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "🤬 Себе потыкай.. Лови бан на минуту 💢",
		})
		return
	}

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
			Text:      fmt.Sprintf("<b>✅ Новый пользователь</b>\n\n<blockquote>id: %d</blockquote>\n<blockquote>username: @%s</blockquote>\n<blockquote>firstname: %s</blockquote>\n", userId, username, firstname),
			ParseMode: models.ParseModeHTML,
		})
	}
	sendStart(ctx, b, chatID)
}

var (
	userActivity  = make(map[int64][]time.Time) // Хранение времени действий пользователей
	userSanctions = make(map[int64]time.Time)   // Хранение санкций
	mu            sync.Mutex                    // Мьютекс для конкурентного доступа
)

const (
	MaxCommands    = 9           // Максимум команд за минуту
	SanctionPeriod = time.Minute // Время санкции
	TimeWindow     = time.Minute // Временное окно для проверки
)

func checkUserActivity(chatID int64) bool {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()

	if sanctionEnd, sanctioned := userSanctions[chatID]; sanctioned {
		if now.Before(sanctionEnd) {
			return false
		}
		delete(userSanctions, chatID)
	}

	activity := userActivity[chatID]
	var newActivity []time.Time

	for _, t := range activity {
		if now.Sub(t) <= TimeWindow {
			newActivity = append(newActivity, t)
		}
	}
	newActivity = append(newActivity, now)
	userActivity[chatID] = newActivity
	if len(newActivity) > MaxCommands {
		userSanctions[chatID] = now.Add(SanctionPeriod) // Блокировка на 1 минуту
		delete(userActivity, chatID)                    // Сбрасываем активность
		return false                                    // Пользователь заблокирован
	}
	return true
}
