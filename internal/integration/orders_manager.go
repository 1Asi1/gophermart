package integration

import (
	"context"
	"time"

	"github.com/1Asi1/gophermart/internal/integration/accrual"
	"github.com/1Asi1/gophermart/internal/repository"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	accrualStatusProcessed = "PROCESSED"
)

type Store interface {
	Update(context.Context, repository.Order) error
	UpdateBalance(ctx context.Context, order repository.Order) error
	GetOrdersNumbers(context.Context) ([]repository.Order, error)
}

type OrdersManager struct {
	client *accrual.Client
	store  Store
	log    zerolog.Logger
}

func New(client *accrual.Client, store Store, log zerolog.Logger) OrdersManager {
	return OrdersManager{
		client: client,
		store:  store,
		log:    log,
	}
}

func (o OrdersManager) Sync(ctx context.Context) {
	l := log.With().Str("integration", "sync").Logger()
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		orders, err := o.store.GetOrdersNumbers(ctx)
		if err != nil {
			l.Error().Err(err).Msg("o.store.GetOrdersNumbers")
		}

		result := make([]repository.Order, 0)
		for _, v := range orders {
			data := repository.Order{
				UserID:  v.UserID,
				Number:  v.Number,
				Status:  v.Status,
				Accrual: v.Accrual,
				Checked: v.Checked,
			}

			resp, err := o.client.GetOrder(v.Number)
			if err != nil {
				l.Error().Err(err).Msg("o.client.GetOrder")
				continue
			}

			data.Status = resp.Status
			data.Accrual = resp.Accrual

			if data.Status == accrualStatusProcessed && data.Accrual != nil && !data.Checked {
				data.Checked = true
				err := o.store.UpdateBalance(ctx, data)
				if err != nil {
					l.Error().Err(err).Msg("o.store.UpdateBalance")
					continue
				}
			}

			result = append(result, data)
		}

		for _, v := range result {
			err := o.store.Update(ctx, v)
			if err != nil {
				l.Error().Err(err).Msg("o.store.Update")
			}
		}
	}
}
