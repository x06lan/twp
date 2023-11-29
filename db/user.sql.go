// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: user.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addUser = `-- name: AddUser :exec

INSERT INTO
    "user" (
        "username",
        "password",
        "name",
        "email",
        "address",
        "image_id",
        "role",
        "credit_card",
        "enabled"
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        '',
        $5,
        'customer',
        '{}',
        TRUE
    )
`

type AddUserParams struct {
	Username string      `json:"username"`
	Password string      `json:"password"`
	Name     string      `json:"name"`
	Email    string      `json:"email"`
	ImageID  pgtype.UUID `json:"image_id"`
}

func (q *Queries) AddUser(ctx context.Context, arg AddUserParams) error {
	_, err := q.db.Exec(ctx, addUser,
		arg.Username,
		arg.Password,
		arg.Name,
		arg.Email,
		arg.ImageID,
	)
	return err
}

const findUserPassword = `-- name: FindUserPassword :one

SELECT "password" FROM "user" WHERE "username" = $1 OR "email" = $1
`

func (q *Queries) FindUserPassword(ctx context.Context, username string) (string, error) {
	row := q.db.QueryRow(ctx, findUserPassword, username)
	var password string
	err := row.Scan(&password)
	return password, err
}

const userExists = `-- name: UserExists :one

SELECT EXISTS (
        SELECT 1
        FROM "user"
        WHERE
            "username" = $1
            OR "email" = $2
    )
`

type UserExistsParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (q *Queries) UserExists(ctx context.Context, arg UserExistsParams) (bool, error) {
	row := q.db.QueryRow(ctx, userExists, arg.Username, arg.Email)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}
