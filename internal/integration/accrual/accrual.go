package accrual

import (
	"fmt"
	"net/http"

	"github.com/1Asi1/gophermart/internal/config"
	"github.com/1Asi1/gophermart/internal/oops"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type Request struct {
	Order string  `json:"order"`
	Goods []Goods `json:"goods"`
}

type Goods struct {
	Description string `json:"description"`
	Price       int    `json:"price"`
}

type Response struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float32 `json:"accrual,omitempty"`
}

type Client struct {
	http *resty.Client
	cfg  config.Config
	log  zerolog.Logger
}

func New(cfg config.Config, log zerolog.Logger) Client {
	client := resty.New()
	return Client{http: client, cfg: cfg, log: log}
}

func (c Client) GetOrder(num string) (Response, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.cfg.AccrualAddr, num)
	var result Response
	request := c.http.R().SetResult(&result)
	request.Method = resty.MethodGet
	request.URL = url

	defer c.http.SetCloseConnection(true)
	resp, err := request.Send()
	if err != nil {
		return Response{}, fmt.Errorf(":%w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return Response{}, fmt.Errorf(":%w", oops.ErrStatusNotOK)
	}

	return result, nil
}
