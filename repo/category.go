package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Category struct {
	Id     int
	UserId int
	Name   string
}

func CreateCategory(ctx context.Context, conn *pgx.Conn, category Category) error {
	sqlQuery := `
	INSERT INTO categories(user_id, name)
	VALUES($1, $2)
	`
	_, err := conn.Exec(ctx, sqlQuery, category.UserId, category.Name)
	return err
}

func GetCategories(ctx context.Context, conn *pgx.Conn, userId int) ([]Category, error) {
	sqlQuery := `
	SELECT id, name
	FROM categories
	WHERE user_id = $1
	`
	rows, err := conn.Query(ctx, sqlQuery, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	categories := make([]Category, 0, 15)
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.Id, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

func GetCategoryByName(ctx context.Context, conn *pgx.Conn, userId int, name string) (Category, error) {
	sqlQuery := `
	SELECT id, name
	FROM categories
	WHERE user_id = $1 AND name = $2
	LIMIT 1
	`
	var c Category
	err := conn.QueryRow(ctx, sqlQuery, userId, name).Scan(&c.Id, &c.Name)
	return c, err
}
