-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN uuid UUID DEFAULT gen_random_uuid() UNIQUE;
UPDATE users SET uuid = gen_random_uuid() WHERE uuid IS NULL;
ALTER TABLE users ALTER COLUMN uuid SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN uuid;
-- +goose StatementEnd