alter table passengers
  add column if not exists document_type text;

update passengers
set document_type = case
  when document is null or btrim(document) = '' then null
  when regexp_replace(document, '[^0-9]', '', 'g') ~ '^[0-9]{11}$' then 'CPF'
  else 'RG'
end
where document_type is null;

alter table booking_passengers
  add column if not exists document_type text;

update booking_passengers
set document_type = case
  when document is null or btrim(document) = '' then null
  when regexp_replace(document, '[^0-9]', '', 'g') ~ '^[0-9]{11}$' then 'CPF'
  else 'RG'
end
where document_type is null;
