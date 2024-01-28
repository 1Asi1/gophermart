CREATE TABLE users (
id uuid primary key ,
login text unique ,
password text,
token text
);