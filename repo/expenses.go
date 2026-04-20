package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Operation struct {
	Id        int
	UserID    int
	Amount    float64
	Category  string
	CreatedAt time.Time
}

type CategoryStat struct {
	Category string
	Total    float64
}

func CreateExpense(ctx context.Context, conn *pgx.Conn, operation Operation) error {
	sqlQuery := `
	INSERT INTO operations (user_id, amount, category)
	VALUES ($1, $2, $3)
	`
	_, err := conn.Exec(ctx, sqlQuery, operation.UserID, operation.Amount, operation.Category)
	return err
}

// Get expenses by categories
func GetExpensesByCategory(ctx context.Context, conn *pgx.Conn, userId int, start time.Time, end time.Time) ([]CategoryStat, error) {
	sqlQuery := `
	SELECT category, SUM(amount) AS TOTAL
	FROM operations
	WHERE user_id = $1
	AND created_at >= $2
	AND created_at < $3
	GROUP BY category
	`

	rows, err := conn.Query(ctx, sqlQuery, userId, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	categorySum := make([]CategoryStat, 0)
	for rows.Next() {
		var c CategoryStat
		if err := rows.Scan(&c.Category, &c.Total); err != nil {
			return nil, err
		}
		categorySum = append(categorySum, c)
	}
	return categorySum, nil
}

// Get All Expenses for month
func GetTotalExpenses(ctx context.Context, conn *pgx.Conn, userId int, start time.Time, end time.Time) (float64, error) {
	sqlQuery := `
	SELECT COALESCE(SUM(amount), 0)
	FROM operations
	WHERE user_id = $1
	AND created_at >= $2
	AND created_at < $3
	`
	row := conn.QueryRow(ctx, sqlQuery, userId, start, end)

	var total float64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// Only 1 category
func GetCategoryExpenses(ctx context.Context, conn *pgx.Conn, userId int, category string) (float64, error) {
	sqlQuery := `
	SELECT COALESCE(SUM(amount), 0) AS TOTAL
	FROM operations
	WHERE user_id = $1
	AND category = $2
	AND created_at >= date_trunc('month', now())
	`
	row := conn.QueryRow(ctx, sqlQuery, userId, category)
	var total float64
	err := row.Scan(&total)
	return total, err
}

func DelteExpense(ctx context.Context, conn *pgx.Conn, userId int) (int64, error) {
	sqlQuery := `
	DELETE FROM operations
	WHERE id = (
	SELECT id FROM operations
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT 1
	)
	`
	result, err := conn.Exec(ctx, sqlQuery, userId)
	if err != nil {
		return 0, err
	}
	rows := result.RowsAffected()
	return rows, nil
}

func History(ctx context.Context, conn *pgx.Conn, userId int) ([]Operation, error) {
	sqlQuery := `
	SELECT amount, category, created_at
	FROM operations
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT 10
	`
	rows, err := conn.Query(ctx, sqlQuery, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	oper := make([]Operation, 0, 10)
	for rows.Next() {
		var o Operation
		if err := rows.Scan(&o.Amount, &o.Category, &o.CreatedAt); err != nil {
			return nil, err
		}
		oper = append(oper, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return oper, nil
}
