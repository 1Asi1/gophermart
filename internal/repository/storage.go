package repository

import (
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/1Asi1/gophermart/internal/oops"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"
)

type User struct {
	ID       uuid.UUID `db:"id"`
	Login    string    `db:"login"`
	Password string    `db:"password"`
	Token    string    `db:"token"`
}

type Order struct {
	UserID     uuid.UUID `db:"user_id"`
	Number     string    `db:"number"`
	Status     string    `db:"status"`
	Accrual    *float32  `db:"accrual"`
	UploadedAt time.Time `db:"uploaded_at"`
	Checked    bool      `db:"checked"`
}

type Balance struct {
	UserID    uuid.UUID `db:"user_id"`
	Current   float32   `db:"current"`
	Withdrawn float32   `db:"withdrawn"`
}

type Withdrawals struct {
	UserID      uuid.UUID `db:"user_id"`
	Number      string    `db:"number"`
	Sum         float32   `db:"sum"`
	ProcessedAt time.Time `db:"processed_at"`
}

type Config struct {
	ConnDSN         string
	MaxConn         int
	MaxConnLifeTime time.Duration
	MaxConnIdleTime time.Duration
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

type Store struct {
	*sqlx.DB
}

func New(cfg Config) (Store, error) {
	db, err := sqlx.Connect("pgx", cfg.ConnDSN)
	if err != nil {
		return Store{}, fmt.Errorf("sqlx.Connect :%w", err)
	}
	db.SetMaxOpenConns(cfg.MaxConn)
	db.SetConnMaxIdleTime(cfg.MaxConnIdleTime)
	db.SetConnMaxLifetime(cfg.MaxConnLifeTime)
	if err = db.Ping(); err != nil {
		return Store{}, fmt.Errorf("db.Ping :%w", err)
	}

	if err = runMigrations(cfg.ConnDSN); err != nil {
		return Store{}, fmt.Errorf("runMigrations :%w", err)
	}

	return Store{db}, nil
}

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err = m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func (s Store) Register(ctx context.Context, user User) error {
	query := `
	INSERT INTO users(id,login,password,token)
	VALUES (:id,:login,:password,:token)`

	resRegister, err := s.NamedExecContext(ctx, query, &user)
	if err != nil {
		return fmt.Errorf("s.NamedExecContext: %w", err)
	}

	if _, err = resRegister.RowsAffected(); err != nil {
		return fmt.Errorf("resRegister.RowsAffected(): %w", err)
	}

	query = `
	INSERT INTO balances(user_id)
	VALUES ($1)`

	resBalance, err := s.ExecContext(ctx, query, user.ID)
	if err != nil {
		return fmt.Errorf("s.ExecContext: %w", err)
	}

	if _, err = resBalance.RowsAffected(); err != nil {
		return fmt.Errorf("resBalance.RowsAffected: %w", err)
	}

	return nil
}

func (s Store) Login(ctx context.Context, user User) (string, error) {
	query := `
	SELECT
	    token
	FROM users
	WHERE login=$1 AND password=$2`

	rows, err := s.QueryContext(ctx, query, user.Login, user.Password)
	if err != nil {
		return "", fmt.Errorf("s.QueryContext: %w", err)
	}
	err = rows.Err()
	if err != nil {
		return "", fmt.Errorf("rows.Err: %w", err)
	}

	var token string
	for rows.Next() {
		err = rows.Scan(&token)
		if err != nil {
			return "", fmt.Errorf("rows.Scan: %w", err)
		}
	}
	defer func() {
		err = rows.Close()
	}()

	return token, nil
}

func (s Store) CheckToken(ctx context.Context, token string) (uuid.UUID, error) {
	query := `
	SELECT
	    id
	FROM users
	WHERE token=$1`

	var id []uuid.UUID
	err := s.SelectContext(ctx, &id, query, token)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("s.SelectContext: %w", err)
	}

	if id == nil {
		return uuid.UUID{}, oops.ErrInvalidToken
	}

	return id[0], nil
}

func (s Store) CreateOrder(ctx context.Context, order Order) error {
	queryChekOrder := `
	SELECT
	    user_id,
		number
	FROM orders
	WHERE user_id=$1 and number=$2
	OR number=$2`

	row := s.QueryRowContext(ctx, queryChekOrder, order.UserID, order.Number)
	err := row.Err()
	if err != nil {
		return fmt.Errorf("s.QueryRowContext: %w", err)
	}

	var chek Order
	_ = row.Scan(&chek.UserID, &chek.Number)

	if chek.UserID == order.UserID && chek.Number == order.Number {
		return oops.ErrOrderCreate
	}

	if chek.UserID != order.UserID && chek.Number == order.Number {
		return oops.ErrOrderReady
	}

	query := `
	INSERT INTO orders(user_id,number,status,accrual,uploaded_at,checked)
	VALUES (:user_id,:number,:status,:accrual,:uploaded_at,:checked)`

	rows, err := s.NamedExecContext(ctx, query, &order)
	if err != nil {
		return fmt.Errorf("s.NamedExecContext: %w", err)
	}

	if _, err = rows.RowsAffected(); err != nil {
		return fmt.Errorf("rows.RowsAffected: %w", err)
	}

	return nil
}

func (s Store) Orders(ctx context.Context, id uuid.UUID) ([]Order, error) {
	query := `
	SELECT
	    number,
	    status,
	    accrual,
	    uploaded_at
	FROM orders
	WHERE user_id=$1`

	var orders []Order
	err := s.SelectContext(ctx, &orders, query, id)
	if err != nil {
		return nil, fmt.Errorf("s.SelectContext: %w", err)
	}

	if orders == nil {
		return nil, oops.ErrEmptyData
	}

	return orders, nil
}

func (s Store) Order(ctx context.Context, id uuid.UUID, number string) (Order, error) {
	query := `
	SELECT
	    accrual
	FROM orders
	WHERE user_id=$1 and number=$2`

	var orders []Order
	err := s.SelectContext(ctx, &orders, query, id, number)
	if err != nil {
		return Order{}, fmt.Errorf("s.SelectContext: %w", err)
	}

	if orders == nil {
		return Order{}, oops.ErrEmptyData
	}

	return orders[0], nil
}

func (s Store) Balance(ctx context.Context, id uuid.UUID) (Balance, error) {
	query := `
	SELECT
	    current,
	    withdrawn
	FROM balances
	WHERE user_id=$1`

	rows, err := s.QueryContext(ctx, query, id)
	if err != nil {
		return Balance{}, fmt.Errorf("s.QueryContext: %w", err)
	}

	err = rows.Err()
	if err != nil {
		return Balance{}, fmt.Errorf("rows.Err: %w", err)
	}

	var balance Balance
	for rows.Next() {
		err = rows.Scan(&balance.Current, &balance.Withdrawn)
		if err != nil {
			return Balance{}, fmt.Errorf("rows.Next: %w", err)
		}
	}
	defer func() {
		err = rows.Close()
	}()

	return balance, nil
}

func (s Store) Withdraw(ctx context.Context, req Order, sum float32) error {
	txBalance, err := s.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("s.Begin: %w", err)
	}

	txWithdraw, err := s.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("s.Begin: %w", err)
	}

	queryBalanceUpdate := `
	UPDATE balances
	SET
	   current=$1,
	   withdrawn=withdrawn+$2
	WHERE user_id=$3`
	resBalanceUpdate, err := txBalance.Exec(queryBalanceUpdate, *req.Accrual, sum, req.UserID)
	if err != nil {
		err = txBalance.Rollback()
		if err != nil {
			return fmt.Errorf("tx.Rollback: %w", err)
		}
		return fmt.Errorf("tx.QueryContext: %w", err)
	}
	_, err = resBalanceUpdate.RowsAffected()
	if err != nil {
		err = txBalance.Rollback()
		if err != nil {
			return fmt.Errorf("tx.Rollback: %w", err)
		}
		return fmt.Errorf("resBalanceUpdate.RowsAffected: %w", err)
	}

	queryWithdrawUpdate := `
	INSERT INTO withdrawns (user_id, number, sum, processed_at)
	VALUES ($1, $2, $3, NOW())
	`
	resWithdrawUpdate, err := txWithdraw.Exec(queryWithdrawUpdate, req.UserID, req.Number, sum)
	if err != nil {
		err = txWithdraw.Rollback()
		if err != nil {
			return fmt.Errorf("tx.Rollback: %w", err)
		}
		return fmt.Errorf("tx.QueryContext: %w", err)
	}
	_, err = resWithdrawUpdate.RowsAffected()
	if err != nil {
		err = txWithdraw.Rollback()
		if err != nil {
			return fmt.Errorf("tx.Rollback: %w", err)
		}
		return fmt.Errorf("resWithdrawUpdate.RowsAffected: %w", err)
	}
	err = txBalance.Commit()
	if err != nil {
		err = txBalance.Rollback()
		if err != nil {
			return fmt.Errorf("tx.Rollback: %w", err)
		}
		return fmt.Errorf("tx.Commit: %w", err)
	}
	err = txWithdraw.Commit()
	if err != nil {
		err = txWithdraw.Rollback()
		if err != nil {
			return fmt.Errorf("tx.Rollback: %w", err)
		}
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (s Store) Withdrawals(ctx context.Context, id uuid.UUID) ([]Withdrawals, error) {
	query := `
	SELECT
	    number,
	    sum,
	    processed_at
	FROM withdrawns
	WHERE user_id=$1`

	var withdrawals []Withdrawals
	err := s.SelectContext(ctx, &withdrawals, query, id)
	if err != nil {
		return nil, fmt.Errorf("s.SelectContext: %w", err)
	}

	if withdrawals == nil {
		return nil, oops.ErrEmptyData
	}

	return withdrawals, nil
}
