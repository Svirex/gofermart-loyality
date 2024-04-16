CREATE TABLE public.users (
    id SERIAL PRIMARY KEY,
    login varchar(128),
    hash text
)