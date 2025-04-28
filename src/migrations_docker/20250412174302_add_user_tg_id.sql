-- +goose Up
-- +goose StatementBegin
alter table appuser add column tg_id text;

-- +goose StatementEnd
