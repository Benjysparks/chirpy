-- +goose Up
ALTER TABLE users ADD hashed_password TEXT DEFAULT 'not set';

-- +goose Down
ALTER TABLE users DROP COLUMN hashed_password;