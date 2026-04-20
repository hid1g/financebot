package bot

import (
	"context"
	"financebot/repo"
	"financebot/utils"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
)

var userState = make(map[int64]string)
var tempAmount = make(map[int64]float64)

func StartBot(bot *tgbotapi.BotAPI, conn *pgx.Conn) {
	ctx := context.Background()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		userId := update.Message.From.ID
		if state, ok := userState[userId]; ok {
			if state == "waiting_amount" {
				text := update.Message.Text
				amount, err := strconv.Atoi(text)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ведите число")
					bot.Send(msg)
					continue
				}
				tempAmount[userId] = float64(amount)
				userState[userId] = "waiting_category"
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Введите категорию"))
				continue
			}
			if state == "waiting_category" {
				category := update.Message.Text
				amount := tempAmount[userId]
				user, err := repo.GetUserByTgId(ctx, conn, userId)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Пользователь не найден"))
					continue
				}
				operation := repo.Operation{
					UserID:   user.Id,
					Amount:   amount,
					Category: category,
				}
				if err := repo.CreateExpense(ctx, conn, operation); err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка создания траты"))
					continue
				}
				delete(userState, userId)
				delete(tempAmount, userId)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Трата добавлена"))
				continue
			}
		}
		switch {
		case strings.HasPrefix(update.Message.Text, "/start"):
			if err := repo.CreateUser(ctx, conn, update.Message.From.ID); err != nil {
				fmt.Println("CreateUser error:", err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка регистрации")
				bot.Send(msg)
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Регистрация выполнена")
			bot.Send(msg)

		case strings.HasPrefix(update.Message.Text, "/add"):
			userID := update.Message.From.ID
			userState[userID] = "waiting_amount"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите сумму:")
			bot.Send(msg)

		case strings.HasPrefix(update.Message.Text, "/month"):
			start, end := utils.CurrentMonth()
			MonthStat(bot, conn, update, start, end)

		case strings.HasPrefix(update.Message.Text, "/last_month"):
			start, end := utils.LastMonth()
			MonthStat(bot, conn, update, start, end)

		case strings.HasPrefix(update.Message.Text, "/deleteLast"):
			DeleteExpenseCMD(bot, conn, update)

		case strings.HasPrefix(update.Message.Text, "/history"):
			ShowHistory(bot, conn, update)
		}
	}
}

func AddTask(bot *tgbotapi.BotAPI, conn *pgx.Conn, update tgbotapi.Update) {
	ctx := context.Background()
	text := update.Message.Text
	pars := strings.Split(text, " ")

	if len(pars) != 3 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "неверный формат команды. формат: /add категория сумма")
		bot.Send(msg)
		return
	}

	category := pars[1]
	amountSTR := pars[2]
	amount, err := strconv.Atoi(amountSTR)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная сумма")
		bot.Send(msg)
		return
	}
	tgId := update.Message.From.ID
	user, err := repo.GetUserByTgId(ctx, conn, tgId)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пользователь не найден. Напишите /start")
		bot.Send(msg)
		return
	}
	userId := user.Id

	operation := repo.Operation{
		UserID:   userId,
		Amount:   float64(amount),
		Category: category,
	}
	if err := repo.CreateExpense(ctx, conn, operation); err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при сохранении")
		bot.Send(msg)
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Добавлено %s - %d", category, amount))

	bot.Send(msg)

}

func MonthStat(bot *tgbotapi.BotAPI, conn *pgx.Conn, update tgbotapi.Update, startTime time.Time, endTime time.Time) {
	ctx := context.Background()
	tgId := update.Message.From.ID
	user, err := repo.GetUserByTgId(ctx, conn, tgId)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пользователь не найден. Напишите /start")
		bot.Send(msg)
		return
	}
	userId := user.Id
	amount, err := repo.GetTotalExpenses(ctx, conn, userId, startTime, endTime)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка получения суммы расходов")
		bot.Send(msg)
		return
	}
	categoryStat, err := repo.GetExpensesByCategory(ctx, conn, userId, startTime, endTime)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка получения расходов по категориям")
		bot.Send(msg)
		return
	}
	text := "Статистика расходов: \n\n"
	text += fmt.Sprintf("Всего: %.2f\n\n", amount)
	for _, c := range categoryStat {
		text += fmt.Sprintf("%s: %.2f ₽\n", c.Category, c.Total)
	}
	if len(categoryStat) == 0 {
		text += "Нет расходов за месяц"
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	bot.Send(msg)

}

func DeleteExpenseCMD(bot *tgbotapi.BotAPI, conn *pgx.Conn, update tgbotapi.Update) {
	ctx := context.Background()
	tgId := update.Message.From.ID
	user, err := repo.GetUserByTgId(ctx, conn, tgId)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пользователь не найден. Напишите /start")
		bot.Send(msg)
		return
	}
	userId := user.Id
	result, err := repo.DelteExpense(ctx, conn, userId)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка удаления!")
		bot.Send(msg)
		return
	}
	if result == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет трат")
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Трата удалена")
		bot.Send(msg)
	}
}

func ShowHistory(bot *tgbotapi.BotAPI, conn *pgx.Conn, update tgbotapi.Update) {
	ctx := context.Background()
	tgId := update.Message.From.ID
	user, err := repo.GetUserByTgId(ctx, conn, tgId)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пользователь не найден. Напишите /start")
		bot.Send(msg)
		return
	}
	userId := user.Id
	oper, err := repo.History(ctx, conn, userId)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка вывода истроии")
		bot.Send(msg)
		return
	}
	if len(oper) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет трат")
		bot.Send(msg)
		return
	}
	text := "📊 История:"
	for _, op := range oper {
		text += fmt.Sprintf("%s - %s %.0f ₽\n",
			op.CreatedAt.Format("02.01"),
			op.Category,
			op.Amount)
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	bot.Send(msg)

}
