package funcExcel

import (
	"os"

	"github.com/joho/godotenv"
)

func BotToken() string {
	err := godotenv.Load(".env")

	if err != nil {
		panic(err)
	}
	return os.Getenv("TELEGRAM_TOKEN")
}
