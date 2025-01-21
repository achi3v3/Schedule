package funcExcel

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	Welcome = "Добро пожаловать"
)

func editMessage(ctx context.Context, b *bot.Bot, chatID int64, messageID int, text string, keyboard *models.InlineKeyboardMarkup) {
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        text,
		ReplyMarkup: keyboard,
		ParseMode:   models.ParseModeHTML,
	})
	if err != nil {
		log.Printf("Ошибка изменения сообщения: %v", err)
	}
}

func handleDocument(ctx context.Context, b *bot.Bot, update *models.Update) {
	connStr := "user=postgres password=password sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Ошибка при подключении к базе данных: %s", err)
	}
	defer db.Close()
	doc := update.Message.Document
	caption := update.Message.Caption

	file, err := b.GetFile(ctx, &bot.GetFileParams{
		FileID: doc.FileID,
	})
	if err != nil {
		fmt.Println("Ошибка получения файла:", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Не удалось получить файл.",
		})
		return
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.Token(), file.FilePath)

	saveDir := "uploaded_files"
	err = os.MkdirAll(saveDir, os.ModePerm)
	if err != nil {
		fmt.Println("Ошибка создания папки:", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Не удалось создать папку для сохранения файла.",
		})
		return
	}

	savePath := filepath.Join(saveDir, doc.FileName)

	err = downloadExcelFile(fileURL, savePath)
	if err != nil {
		fmt.Println("Ошибка скачивания файла:", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Не удалось сохранить файл.",
		})
		return
	}

	response := fmt.Sprintf("✅ Успешно сохранено!\nℹ️ Файл: '%s'\n", doc.FileName)
	if caption != "" {
		response += fmt.Sprintf("ℹ️ Установленная неделя: %s", caption)
	}
	ReloadFile(doc.FileName, caption)
	createTableUsers(db)
	createDataBasesExcel(db)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   response,
	})
}

func downloadExcelFile(fileURL, savePath string) error {
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
