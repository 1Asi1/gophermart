package integration

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/1Asi1/gophermart/internal/integration/accrual"
	"github.com/1Asi1/gophermart/internal/oops"
	"github.com/1Asi1/gophermart/internal/repository"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	accrualStatusProcessed = "PROCESSED"
)
const (
	workCounter = 10
	jobCounter  = 100
	timeSleep   = 60
)

type Store interface {
	Update(context.Context, repository.Order) error
	UpdateBalance(ctx context.Context, order repository.Order) error
	GetOrdersNumbers(context.Context, int) ([]repository.Order, error)
}

type worker struct {
	job chan repository.Order
	err chan error
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
	stop := new(atomic.Bool)
	var workers worker
	workers.job = make(chan repository.Order, jobCounter)
	workers.err = make(chan error, 1)

	for i := 0; i < workCounter; i++ {
		go func(w worker) {
			for j := range w.job {
				if !stop.Load() {
					o.orderWork(j, w.err)
				}
			}
		}(workers)
	}

	l := log.With().Str("integration", "sync").Logger()
	ticker := time.NewTicker(1 * time.Second)
	var offset int
	go func() {
		for range workers.err {
			stop.Store(true)
			time.Sleep(timeSleep * time.Second)
			stop.Store(false)
		}
	}()
	for range ticker.C {
		if !stop.Load() {
			orders, err := o.store.GetOrdersNumbers(ctx, offset)
			if err != nil {
				l.Error().Err(err).Msg("o.store.GetOrdersNumbers")
			}
			offset = len(orders)

			for _, j := range orders {
				workers.job <- j
			}
		}
	}
}

func (o OrdersManager) orderWork(order repository.Order, errCh chan error) {
	l := log.With().Str("integration", "orderWork").Logger()

	resp, err := o.client.GetOrder(order.Number)
	if err != nil {
		l.Error().Err(err).Msg("o.client.GetOrder")
		if errors.Is(err, oops.ErrStatusTooManyRequests) {
			errCh <- err
		}
		return
	}

	data := repository.Order{
		UserID:  order.UserID,
		Number:  order.Number,
		Status:  resp.Status,
		Accrual: resp.Accrual,
		Checked: order.Checked,
	}

	if data.Status == accrualStatusProcessed && data.Accrual != nil && !data.Checked {
		data.Checked = true
		err := o.store.UpdateBalance(context.Background(), data)
		if err != nil {
			l.Error().Err(err).Msg("o.store.UpdateBalance")
			return
		}
	}

	err = o.store.Update(context.Background(), data)
	if err != nil {
		l.Error().Err(err).Msg("o.store.Update")
		return
	}
}
