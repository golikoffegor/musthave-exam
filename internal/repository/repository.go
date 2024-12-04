package repository

import (
	"context"
	"database/sql"
	"log"
	"musthave-exam/internal/model"
	"musthave-exam/migrations"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
)

type repo struct {
	db  *sql.DB
	log *logrus.Logger
}

const (
	Debit    = "Debit"
	Withdraw = "Withdraw"
)

type UserRepository interface {
	RegisterUser(ctx context.Context, user model.User) (int64, error)
	LoginUser(ctx context.Context, user model.User) (int64, error)
	GetUser(ctx context.Context, id int64) (model.User, error)
}

type Repository interface {
	RegisterUser(ctx context.Context, user model.User) (int64, error)
	LoginUser(ctx context.Context, user model.User) (int64, error)
	GetUser(ctx context.Context, id int64) (model.User, error)

	AddOrder(ctx context.Context, userID int64, number string) (*model.Transaction, error)
	GetOrders(ctx context.Context, userID int64, act string) ([]model.Transaction, error)
	GetBalance(ctx context.Context, userID int64) (model.Balance, error)
	Withdraw(ctx context.Context, userID int64, order string, sum float64) error

	UpdateTransactionStatus(ctx context.Context, transactionID string, status string) error
	UpdateTransactionStatusAndAccrual(ctx context.Context, transactionID string, status string, accrual float64) error
	GetNewTransactions(ctx context.Context) ([]model.Transaction, error)
}

func NewRepository(dbsql *sql.DB, logger *logrus.Logger) *repo {
	p := &repo{db: dbsql, log: logger}
	err := p.migrateDB()
	if err != nil {
		log.Fatal(err)
	}
	return p
}

func (r *repo) migrateDB() error {
	goose.SetBaseFS(migrations.Migrations)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := goose.RunContext(ctx, "up", r.db, ".")
	if err != nil {
		return err
	}

	return nil
}
