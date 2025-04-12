-- +goose Up
-- +goose StatementBegin
alter table poster add column kp_id text;
alter table poster add column image_url text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table poster drop column kp_id;
alter table poster drop column image_url;
-- +goose StatementEnd
