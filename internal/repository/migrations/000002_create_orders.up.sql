CREATE TABLE orders (
user_id uuid,
number text unique,
status text,
accrual float default 0,
uploaded_at timestamptz,
checked bool
);