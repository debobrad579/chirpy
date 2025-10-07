-- +goose Up
ALTER TABLE users
    ADD COLUMN is_chirpy_red boolean NOT NULL;

-- +goose Down
ALTER TABLE users
    DROP COLUMN is_chirpy_red;

