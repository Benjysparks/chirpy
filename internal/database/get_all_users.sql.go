// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: get_all_users.sql

package database

import (
	"context"
)

const getAllUsers = `-- name: GetAllUsers :many
SELECT id, created_at, updated_at, email, hashed_password, is_chirpy_red, username FROM users
ORDER BY email
`

func (q *Queries) GetAllUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, getAllUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Email,
			&i.HashedPassword,
			&i.IsChirpyRed,
			&i.Username,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
