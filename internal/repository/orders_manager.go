package repository

import (
	"context"
	"fmt"

	"github.com/1Asi1/gophermart/internal/oops"
)

func (s Store) Update(ctx context.Context, order Order) error {
	query := `
	UPDATE orders
	SET accrual=$1,status=$2,checked=$3
	WHERE user_id=$4 AND number=$5`
	rows, err := s.QueryContext(ctx, query, order.Accrual, order.Status, order.Checked, order.UserID, order.Number)
	if err != nil {
		return fmt.Errorf("s.QueryContext: %w", err)
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("rows.Err: %w", err)
	}

	return nil
}

func (s Store) UpdateBalance(ctx context.Context, order Order) error {
	query := `
	UPDATE balances
	SET current=current+$1
	WHERE user_id=$2`
	rows, err := s.QueryContext(ctx, query, *order.Accrual, order.UserID)
	if err != nil {
		return fmt.Errorf("s.QueryContext: %w", err)
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("rows.Err: %w", err)
	}

	return nil
}

func (s Store) GetOrdersNumbers(ctx context.Context) ([]Order, error) {
	query := `
	SELECT
	    user_id,
	    number,
	    status,
	    accrual,
	    uploaded_at,
	    checked
	FROM orders
	WHERE NOT checked`

	var orders []Order
	err := s.SelectContext(ctx, &orders, query)
	if err != nil {
		return nil, fmt.Errorf("s.SelectContext: %w", err)
	}

	if orders == nil {
		return nil, oops.ErrEmptyData
	}

	return orders, nil
}
