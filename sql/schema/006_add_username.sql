-- +goose Up
ALTER TABLE users ADD username TEXT NOT NULL UNIQUE;

-- +goose Down
ALTER TABLE users DROP COLUMN username;