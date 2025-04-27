-- +goose Up
-- +goose StatementBegin
drop index historyrecord_poster_id_idx;
drop index poster_kp_id_idx;

-- get rid of Kinopoisk
alter table poster drop column kp_id;
alter table poster drop column rating;

alter table HistoryRecord rename to PosterRecord;

-- +goose StatementEnd


