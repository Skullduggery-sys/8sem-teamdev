-- +goose Up
-- +goose StatementBegin
create function adopt_list_childs()
returns trigger as $$
begin
    update listposter
    set list_id=old.parent_id
    where list_id=old.id;

    update list
    set parent_id=old.parent_id
    from (select lc.id from list lp join list lc on lp.id = lc.parent_id where lp.id = old.id) as child
    where list.id=child.id;

    return old;
end;
$$ language plpgsql;

create trigger adopt_list_childs_trigger
before delete on list
for each row
execute function adopt_list_childs();
-- +goose StatementEnd


