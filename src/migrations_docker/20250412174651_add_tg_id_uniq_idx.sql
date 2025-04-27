-- +goose NO TRANSACTION
-- +goose Up
create unique index concurrently user_id_tg_id_uniq_index on appuser(id, tg_id);
