-- +goose NO TRANSACTION
-- +goose Up
create unique index concurrently on appuser(login);

create unique index concurrently on poster(kp_id);

create unique index concurrently list_id_parentid_unique_index on list(id, parent_id);

create unique index concurrently listposter_list_poster_ids_unique_index on listposter(list_id, poster_id);

create unique index concurrently on historyrecord(poster_id);

