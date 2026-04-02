alter table if exists stops
  add column if not exists latitude numeric(9,6),
  add column if not exists longitude numeric(9,6);
