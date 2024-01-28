package models

import (
	"errors"
)

type UserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u UserRequest) Validate() error {
	if u.Login == "" || u.Password == "" {
		return errors.New("empty login or password")
	}

	return nil
}
