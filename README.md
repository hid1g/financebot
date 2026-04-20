Finance Telegram Bot

A simple Telegram bot for tracking personal expenses.
Built as a first backend project using Go.

Features

* Add expenses using a step-by-step dialog
* View statistics for the current and previous month
* See recent transactions
* Delete the last expense

Commands

/start        register user
/add          add new expense
/month        current month statistics
/last_month   previous month statistics
/history      recent transactions
/deleteLast   delete last expense

How /add works

/add
→ enter amount
→ 500
→ enter category
→ food
→ expense saved

Setup

Clone the repository:

git clone https://github.com/your-username/finance-bot.git
cd finance-bot

Create a .env file:

DATABASE_URL=postgres://postgres:password@localhost:5432/postgres
BOT_TOKEN=your_bot_token

Run the bot:

go run main.go

Project structure

bot/         telegram handlers
repo/        database queries
connection/  database connection
utils/       helper functions
main.go      entry point

Stack

* Go
* PostgreSQL
* pgx
* Telegram Bot API
* migrate
