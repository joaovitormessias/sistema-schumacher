-- 0002_seed_roles.sql
insert into roles (name) values ('admin') on conflict do nothing;
insert into roles (name) values ('operador') on conflict do nothing;
insert into roles (name) values ('financeiro') on conflict do nothing;
