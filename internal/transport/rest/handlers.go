package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/1Asi1/gophermart/internal/models"
	"github.com/1Asi1/gophermart/internal/oops"
	"github.com/1Asi1/gophermart/internal/service"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type handlers struct {
	service service.Service
	log     zerolog.Logger
}

func newHandlers(s service.Service, log zerolog.Logger) handlers {
	return handlers{service: s, log: log}
}

func (h *handlers) register(w http.ResponseWriter, r *http.Request) {
	l := h.log.With().Str("route", "register").Logger()

	var user models.UserRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		l.Error().Err(err).Msg("json.NewDecoder")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = user.Validate(); err != nil {
		l.Error().Err(err).Msg("user.Validate")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.service.Register(r.Context(), user)
	if err != nil {
		l.Error().Err(err).Msg(" h.service.Register")
		if errors.Is(err, errors.New("занят")) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}

func (h *handlers) login(w http.ResponseWriter, r *http.Request) {
	l := h.log.With().Str("route", "login").Logger()

	var user models.UserRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		l.Error().Err(err).Msg("json.NewDecoder")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = user.Validate(); err != nil {
		l.Error().Err(err).Msg("user.Validate")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(r.Context(), user)
	if err != nil {
		l.Error().Err(err).Msg("h.service.Login")
		if errors.Is(err, errors.New("неверная пара логин/пароль")) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}

func (h *handlers) createOrder(w http.ResponseWriter, r *http.Request) {
	l := h.log.With().Str("route", "createOrder").Logger()

	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		l.Error().Msg("Content-Type invalid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var num int
	err := json.NewDecoder(r.Body).Decode(&num)
	if err != nil {
		l.Error().Err(err).Msg("json.NewDecoder")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(r.Header.Get("ID"))
	if err != nil {
		l.Error().Err(err).Msg("uuid.Parse")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := models.OrderRequest{UserID: id, Number: strconv.Itoa(num)}
	if err = req.Validate(); err != nil {
		l.Error().Err(err).Msg("req.Validate()")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	l.Info().Msgf("value: %+v", req)

	err = h.service.CreateOrder(r.Context(), req)
	if err != nil {
		l.Error().Err(err).Msg("h.service.CreateOrder")
		if errors.Is(err, oops.ErrOrderReady) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		if errors.Is(err, oops.ErrOrderCreate) {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *handlers) getOrders(w http.ResponseWriter, r *http.Request) {
	l := h.log.With().Str("route", "getOrders").Logger()

	id, err := uuid.Parse(r.Header.Get("ID"))
	if err != nil {
		l.Error().Err(err).Msg("uuid.Parse key: ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := h.service.Orders(r.Context(), id)
	if err != nil {
		l.Error().Err(err).Msg("h.service.Orders")
		if errors.Is(err, oops.ErrEmptyData) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(data)
	if err != nil {
		l.Error().Err(err).Msg("json.Marshal")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	if err != nil {
		l.Error().Err(err).Msg("w.Write")
	}
}

func (h *handlers) getBalance(w http.ResponseWriter, r *http.Request) {
	l := h.log.With().Str("route", "getBalance").Logger()

	id, err := uuid.Parse(r.Header.Get("ID"))
	if err != nil {
		l.Error().Err(err).Msg("uuid.Parse key: ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := h.service.Balance(r.Context(), id)
	if err != nil {
		l.Error().Err(err).Msg("h.service.Balance")
		if errors.Is(err, errors.New("пустой")) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(data)
	if err != nil {
		l.Error().Err(err).Msg("json.Marshal")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	if err != nil {
		l.Error().Err(err).Msg("w.Write")
	}
}

func (h *handlers) withdraw(w http.ResponseWriter, r *http.Request) {
	l := h.log.With().Str("route", "withdraw").Logger()

	var req models.WithdrawRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		l.Error().Err(err).Msg("json.NewDecoder")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	id, err := uuid.Parse(r.Header.Get("ID"))
	if err != nil {
		l.Error().Err(err).Msg("uuid.Parse")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.service.Withdraw(r.Context(), id, req)
	if err != nil {
		l.Error().Err(err).Msg("h.service.Withdraw")
		if errors.Is(err, oops.ErrEmptyData) {
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.Is(err, oops.ErrOrderNumberInvalid) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if errors.Is(err, oops.ErrInsufficientFunds) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handlers) getWithdrawals(w http.ResponseWriter, r *http.Request) {
	l := h.log.With().Str("route", "getWithdrawals").Logger()

	id, err := uuid.Parse(r.Header.Get("ID"))
	if err != nil {
		l.Error().Err(err).Msg("uuid.Parse key: ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := h.service.Withdrawals(r.Context(), id)
	if err != nil {
		l.Error().Err(err).Msg("h.service.Withdrawals")
		if errors.Is(err, oops.ErrEmptyData) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(data)
	if err != nil {
		l.Error().Err(err).Msg("json.Marshal")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	if err != nil {
		l.Error().Err(err).Msg("w.Write")
	}
}
