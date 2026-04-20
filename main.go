package main

import (
	"context"
	"financebot/bot"
	"financebot/connection"
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	godotenv.Load()
	fmt.Println("DB:", os.Getenv("DATABASE_URL"))
	conn, err := connection.CreateConnection(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		panic("BOT_TOKEN NOT SET")
	}

	botApi, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	commands := []tgbotapi.BotCommand{

		{Command: "start", Description: "Начать работу"},
		{Command: "add", Description: "Добавить трату"},
		{Command: "month", Description: "Статистика за месяц"},
		{Command: "last_month", Description: "Статистика за прошлый месяц"},
		{Command: "history", Description: "История последних 10 трат"},
		{Command: "delete", Description: "Удалить последнюю трату"},
	}

	if _, err := botApi.Request(tgbotapi.NewSetMyCommands(commands...)); err != nil {
		panic(err)
	}
	bot.StartBot(botApi, conn)
	fmt.Println("DB:", os.Getenv("DATABASE_URL"))
}
