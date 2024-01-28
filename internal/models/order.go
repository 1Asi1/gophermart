package models

import (
	"strconv"
	"time"

	"github.com/1Asi1/gophermart/internal/oops"
	"github.com/google/uuid"
)

type OrderRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Number string    `json:"number"`
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float32  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (req *OrderRequest) Validate() error {
	ok := luhnAlgorithm(req.Number)
	if !ok {
		return oops.ErrLuhnValidate
	}

	return nil
}

func luhnAlgorithm(number string) bool {
	sum := 0
	isSecondDigit := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))

		if isSecondDigit {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecondDigit = !isSecondDigit
	}

	return sum%10 == 0
}
