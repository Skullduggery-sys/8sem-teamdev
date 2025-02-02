-- +goose Up
-- +goose StatementBegin
create type role_type as enum ('user', 'admin');

create table AppUser (
    id serial primary key,
    name text not null,
    login text not null,
    role role_type default 'user',
    password text not null
);

create table Poster (
    id serial primary key,
    kp_id text not null,
    rating real,
    name text not null,
    genres text,
    year int not null,
    chrono int not null,
    user_id int not null references appuser(id),
    created_at timestamp default now()
);

create table List (
    id serial primary key,
    name text not null,
    user_id int references appuser(id),
    parent_id int references List,
    is_root boolean default false
);

create table ListPoster (
    id serial primary key,
    list_id int not null,
    poster_id int not null,
    position int not null,
    foreign key (list_id) references list(id),
    foreign key (poster_id) references poster(id)
);

create table HistoryRecord (
    id serial primary key,
    poster_id int not null,
    user_id int not null references appuser(id),
    foreign key (poster_id) references poster(id),
    created_at timestamp default now()
);

insert into AppUser (id, name, login, role, password) values (0, 'zero-admin', 'zero-admin', 'admin', 'zero-admin');

insert into List (id, parent_id, name, user_id, is_root) values (default, null, 'root', 0, true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table ListPoster cascade;
drop table HistoryRecord;
drop table Poster;
drop table List;
drop table AppUser;
drop type role_type;
-- +goose StatementEnd
