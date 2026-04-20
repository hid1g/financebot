package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type User struct {
	Id        int
	Tgid      int64
	Createdat time.Time
}

func CreateUser(ctx context.Context, conn *pgx.Conn, tgId int64) error {
	sqlQuerry := `
	INSERT INTO users (telegram_id)
	VALUES ($1)
	ON CONFLICT (telegram_id) DO NOTHING
	`
	_, err := conn.Exec(ctx, sqlQuerry, tgId)
	return err
}

func GetUserByTgId(ctx context.Context, conn *pgx.Conn, tgId int64) (User, error) {
	sqlQuerry := `
	SELECT id
		FROM users
		WHERE telegram_id = $1
	`
	row := conn.QueryRow(ctx, sqlQuerry, tgId)
	var user User
	if err := row.Scan(&user.Id); err != nil {
		return User{}, err
	}
	return user, nil
}
