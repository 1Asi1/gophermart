package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/1Asi1/gophermart/internal/integration/accrual"
	"github.com/1Asi1/gophermart/internal/models"
	"github.com/1Asi1/gophermart/internal/oops"
	"github.com/1Asi1/gophermart/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type Store interface {
	Register(context.Context, repository.User) error
	Login(context.Context, repository.User) (string, error)
	CheckToken(context.Context, string) (uuid.UUID, error)
	CreateOrder(context.Context, repository.Order) error
	Order(context.Context, uuid.UUID, string) (repository.Order, error)
	Orders(context.Context, uuid.UUID) ([]repository.Order, error)
	Balance(context.Context, uuid.UUID) (repository.Balance, error)
	Withdraw(context.Context, repository.Order, float32) error
	Withdrawals(context.Context, uuid.UUID) ([]repository.Withdrawals, error)
}

type Service struct {
	store  Store
	client accrual.Client
}

func New(store Store, client accrual.Client) Service {
	return Service{store: store, client: client}
}

func (s *Service) Register(ctx context.Context, u models.UserRequest) (string, error) {
	pass := getHashPassword(u)

	token := getToken(u.Login, pass)

	model := repository.User{
		ID:       uuid.New(),
		Login:    u.Login,
		Password: pass,
		Token:    token,
	}
	if err := s.store.Register(ctx, model); err != nil {
		return "", fmt.Errorf(":%w", err)
	}

	return token, nil
}

func (s *Service) Login(ctx context.Context, u models.UserRequest) (string, error) {
	pass := getHashPassword(u)

	model := repository.User{
		Login:    u.Login,
		Password: pass,
	}

	token, err := s.store.Login(ctx, model)
	if err != nil {
		return "", fmt.Errorf(":%w", err)
	}

	return token, nil
}

func (s *Service) CheckAccess(ctx context.Context, token string) (string, error) {
	id, err := s.store.CheckToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("s.store.CheckToken: %w", err)
	}

	return id.String(), nil
}

func (s *Service) CreateOrder(ctx context.Context, req models.OrderRequest) error {
	model := repository.Order{
		UserID:     req.UserID,
		Number:     req.Number,
		Status:     "NEW",
		UploadedAt: time.Now(),
	}

	err := s.store.CreateOrder(ctx, model)
	if err != nil {
		return fmt.Errorf("s.store.CreateOrder :%w", err)
	}

	return nil
}

func (s *Service) Orders(ctx context.Context, id uuid.UUID) ([]models.Order, error) {
	orders, err := s.store.Orders(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(":%w", err)
	}

	result := make([]models.Order, len(orders))
	for i, v := range orders {
		result[i] = models.Order{
			Number:     v.Number,
			Status:     v.Status,
			Accrual:    v.Accrual,
			UploadedAt: v.UploadedAt,
		}
	}

	return result, nil
}

func (s *Service) Balance(ctx context.Context, id uuid.UUID) (models.Balance, error) {
	balance, err := s.store.Balance(ctx, id)
	if err != nil {
		return models.Balance{}, fmt.Errorf(":%w", err)
	}

	return models.Balance{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}, nil
}

func (s *Service) Withdraw(ctx context.Context, id uuid.UUID, req models.WithdrawRequest) error {
	balance, err := s.store.Balance(ctx, id)
	if err != nil {
		return fmt.Errorf("s.store.Balance: %w", err)
	}

	if balance.Current-req.Sum < 0 {
		return oops.ErrInsufficientFunds
	}

	balance.Current -= req.Sum

	model := repository.Order{
		UserID:  id,
		Number:  req.Order,
		Accrual: &balance.Current,
	}

	err = s.store.Withdraw(ctx, model, req.Sum)
	if err != nil {
		return fmt.Errorf(":%w", err)
	}

	return nil
}

func (s *Service) Withdrawals(ctx context.Context, id uuid.UUID) ([]models.Withdraw, error) {
	result, err := s.store.Withdrawals(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(":%w", err)
	}

	withdrawals := make([]models.Withdraw, len(result))
	for i, v := range result {
		withdrawals[i] = models.Withdraw{
			Order:       v.Number,
			Sum:         v.Sum,
			ProcessedAt: v.ProcessedAt,
		}
	}

	return withdrawals, nil
}

func getHashPassword(u models.UserRequest) string {
	hash := sha256.Sum256([]byte(u.Password))
	return hex.EncodeToString(hash[:])
}

func getToken(login, password string) string {
	tokenString := fmt.Sprintf("%s%s", login, password)
	hash := sha256.Sum256([]byte(tokenString))
	token := hex.EncodeToString(hash[:])

	return token
}
