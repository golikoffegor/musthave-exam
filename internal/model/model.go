package model

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrEmptyRequestBody   = errors.New("request body is empty")
	ErrErrorRequestBody   = errors.New("failed to read request body")
	ErrFailedToDecodeJSON = errors.New("failed to decode JSON")
	ErrInternalServer     = errors.New("internal server error")
	ErrNotAuthorized      = errors.New("user not authenticated")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidOrderNumber = errors.New("invalid order number format")
	ErrAddOrder           = errors.New("failed to add order")
	ErrAddExistsOrder     = errors.New("order number already exists")
	ErrInvalidLoginPass   = errors.New("invalid login/password")
	ErrLoginAlreadyTaken  = errors.New("login already taken")
	ErrEmptyResponse      = errors.New("empty response")
	ErrIncFunds           = errors.New("insufficient funds")
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

type User struct {
	ID       int64   `json:"id"`
	Login    string  `json:"login"`
	Password string  `json:"password"`
	Balance  float64 `json:"balance"`
}

type Transaction struct {
	ID     string    `json:"number"`
	UserID int64     `json:"user"`
	Summ   float64   `json:"accrual"`
	Date   time.Time `json:"uploaded_at"`
	Status string    `json:"status"`
	Action string    `json:"action"`
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type Withdraw struct {
	ID   string    `json:"order"`
	Summ float64   `json:"sum"`
	Date time.Time `json:"processed_at"`
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
