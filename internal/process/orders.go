package process

import (
	"context"
	"time"

	"github.com/golikoffegor/musthave-exam/internal/app"
	"github.com/golikoffegor/musthave-exam/internal/model"
	"github.com/golikoffegor/musthave-exam/internal/process/repolistener"
	"github.com/golikoffegor/musthave-exam/internal/repository"

	"github.com/sirupsen/logrus"
)

type order struct {
	repo      repository.Repository
	log       *logrus.Logger
	address   string
	iterTime  time.Duration
	pauseTime time.Duration
	Listener  *repolistener.Listener
}

func NewOrderProcess(storage repository.Repository, logger *logrus.Logger, recalcAddress string) *order {
	return &order{
		iterTime:  1 * time.Second,
		pauseTime: 1 * time.Minute,
		repo:      storage,
		log:       logger,
		address:   recalcAddress,
		Listener:  repolistener.NewListener(),
	}
}

func (o *order) WaitTransactionProcessing(ctx context.Context) {
	for {
		select {
		case orderID := <-o.Listener.NewOrderChan: // Вызов функции для обработки нового заказа
			o.log.Debug("Processing new order with ID: ", orderID)
			go o.processTransaction(ctx, model.Transaction{ID: orderID})
		case <-ctx.Done():
			o.log.Info("Stopping new order processor")
			return
		}
	}
}

func (o *order) StartTransactionProcessing(ctx context.Context, r repository.Repository) {
	o.log.Debug("Запуск отслеживания начисления транзакций")
	transactions, err := o.repo.GetNewTransactions(ctx)
	if err != nil {
		o.log.WithError(err).Warning("Failed to get new transactions")
		return // No Retry after some time
	}

	for _, transaction := range transactions {
		go o.processTransaction(ctx, transaction)
	}
}

func (o *order) processTransaction(ctx context.Context, transaction model.Transaction) {
	for {
		// o.log.Debug("Started processTransaction ", transaction.ID)
		select {
		case <-ctx.Done():
			o.log.Info("Stopping processing for transaction due to context cancellation", transaction.ID)
			return
		default:
			o.log.Debug("Check processTransaction ", transaction.ID)
			accrualResponse, err := app.FetchAccrual(o.address, transaction.ID, o.log)
			if err != nil {
				o.log.WithError(err).Info("Error fetching accrual for transaction")
				time.Sleep(o.pauseTime) // Retry after some time
				continue
			}

			switch accrualResponse.Status {
			case "REGISTERED":
				o.log.Info("REGISTERED ", transaction.ID)
				err = o.repo.UpdateTransactionStatus(ctx, transaction.ID, "PROCESSING")
			case "INVALID":
				o.log.Info("INVALID ", transaction.ID)
				err = o.repo.UpdateTransactionStatus(ctx, transaction.ID, "INVALID")
				if err == nil {
					return // Stop processing
				}
			case "PROCESSING":
				o.log.Info("PROCESSING ", transaction.ID)
				if transaction.Status == "NEW" {
					err = o.repo.UpdateTransactionStatus(ctx, transaction.ID, "PROCESSING")
				}
			case "PROCESSED":
				err = o.repo.UpdateTransactionStatusAndAccrual(ctx, transaction.ID, "PROCESSED", accrualResponse.Accrual)
				if err == nil {
					o.log.Info("PROCESSED ", transaction.ID)
					return // Stop processing
				}
			}

			if err != nil {
				o.log.WithError(err).Warning("Error updating status for transaction", transaction.ID)
			}

			time.Sleep(o.iterTime) // Polling interval
		}
	}

}
