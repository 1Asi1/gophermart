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
	res, err := s.ExecContext(ctx, query, order.Accrual, order.Status, order.Checked, order.UserID, order.Number)
	if err != nil {
		return fmt.Errorf("s.QueryContext: %w", err)
	}

	_, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("res.RowsAffected(): %w", err)
	}

	return nil
}

func (s Store) UpdateBalance(ctx context.Context, order Order) error {
	query := `
	UPDATE balances
	SET current=current+$1
	WHERE user_id=$2`
	res, err := s.ExecContext(ctx, query, *order.Accrual, order.UserID)
	if err != nil {
		return fmt.Errorf("s.QueryContext: %w", err)
	}

	_, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("res.RowsAffected(): %w", err)
	}

	return nil
}

func (s Store) GetOrdersNumbers(ctx context.Context, offset int) ([]Order, error) {
	query := `
	SELECT
	    user_id,
	    number,
	    status,
	    accrual,
	    uploaded_at,
	    checked
	FROM orders
	WHERE NOT checked
	ORDER BY uploaded_at
	LIMIT 100 OFFSET $1
	`

	var orders []Order
	err := s.SelectContext(ctx, &orders, query, offset)
	if err != nil {
		return nil, fmt.Errorf("s.SelectContext: %w", err)
	}

	if orders == nil {
		return nil, oops.ErrEmptyData
	}

	return orders, nil
}
