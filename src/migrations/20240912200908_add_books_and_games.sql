-- +goose Up
-- +goose StatementBegin
drop index historyrecord_poster_id_idx;
drop index poster_kp_id_idx;

-- get rid of Kinopoisk
alter table poster drop column kp_id;
alter table poster drop column rating;

alter table HistoryRecord rename to PosterRecord;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table PosterRecord rename to HistoryRecord;

alter table poster add column rating real;
alter table poster add column kp_id text;

create unique index on historyrecord(poster_id);
create unique index on poster(kp_id);
-- +goose StatementEnd
