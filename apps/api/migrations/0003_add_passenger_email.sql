-- 0003_add_passenger_email.sql
alter table booking_passengers add column if not exists email text;
