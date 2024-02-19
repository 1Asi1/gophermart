CREATE TABLE withdrawns (
user_id uuid ,
number text ,
sum float default 0,
processed_at timestamptz
);