-- +goose Up
-- +goose StatementBegin
alter table appuser add column tg_id text;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table appuser drop column tg_id text;
-- +goose StatementEnd
