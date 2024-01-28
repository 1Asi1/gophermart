CREATE TABLE balances (
user_id uuid primary key ,
current float default 0,
withdrawn float default 0
);