-- +goose Up
-- +goose StatementBegin
alter table AppUser add column email text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table AppUser drop column email;
-- +goose StatementEnd
