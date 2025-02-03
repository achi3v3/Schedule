package functions

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	_ "github.com/lib/pq"
)

func 	NewBot() {
	launchNewBot()
}

func launchNewBot() {
	databaseHandler()

	token := BotToken()

	fmt.Println("▶️ ", token)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(UniversalHandler),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}
	b.RegisterHandlerMatchFunc(isStart, Start)
	b.Start(ctx)
}
