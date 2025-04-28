-- +goose Up
-- +goose StatementBegin
alter table poster add column kp_id text;
alter table poster add column image_url text;
-- +goose StatementEnd
