package process

import (
	"context"
	"musthave-exam/internal/app"
	"musthave-exam/internal/model"
	"musthave-exam/internal/repository"
	"time"

	"github.com/sirupsen/logrus"
)

func NewOrderProcess(storage repository.Repository, logger *logrus.Logger, recalcAddress string) *order {
	return &order{
		iterTime: 1 * time.Second,
		repo:     storage,
		log:      logger,
		address:  recalcAddress,
	}
}

type order struct {
	repo     repository.Repository
	log      *logrus.Logger
	address  string
	iterTime time.Duration
}

func (o *order) StartTransactionProcessing(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			o.log.Debug("Stopping transaction processing due to context cancellation")
			return
		default:
			transactions, err := o.repo.GetNewTransactions(ctx)
			if err != nil {
				o.log.WithError(err).Warning("Failed to get new transactions")
				// o.log.Error("Failed to get new transactions: ", err)
				time.Sleep(1 * time.Minute) // Retry after some time
				continue
			}

			for _, transaction := range transactions {
				go o.processTransaction(ctx, transaction)
			}

			time.Sleep(1 * time.Minute) // Polling interval to check for new transactions
		}
	}
}

func (o *order) processTransaction(ctx context.Context, transaction model.Transaction) {
	for {
		select {
		case <-ctx.Done():
			o.log.Info("Stopping processing for transaction due to context cancellation", transaction.ID)
			return
		default:
			accrualResponse, err := app.FetchAccrual(o.address, transaction.ID, o.log)
			if err != nil {
				o.log.WithError(err).Info("Error fetching accrual for transaction")
				time.Sleep(o.iterTime) // Retry after some time
				continue
			}

			switch accrualResponse.Status {
			case "REGISTERED":
				o.log.Info("REGISTERED", transaction.ID)
				err = o.repo.UpdateTransactionStatus(ctx, transaction.ID, "PROCESSING")
			case "INVALID":
				o.log.Info("INVALID", transaction.ID)
				err = o.repo.UpdateTransactionStatus(ctx, transaction.ID, "INVALID")
				if err == nil {
					return // Stop processing
				}
			case "PROCESSING":
				o.log.Info("PROCESSING", transaction.ID)
				if transaction.Status == "NEW" {
					err = o.repo.UpdateTransactionStatus(ctx, transaction.ID, "PROCESSING")
				}
			case "PROCESSED":
				err = o.repo.UpdateTransactionStatusAndAccrual(ctx, transaction.ID, "PROCESSED", accrualResponse.Accrual)
				if err == nil {
					o.log.Info("PROCESSED", transaction.ID)
					return // Stop processing
				}
			}

			if err != nil {
				o.log.WithError(err).Warning("Error updating status for transaction", transaction.ID)
			}

			time.Sleep(1 * time.Minute) // Polling interval
		}
	}
}

// func StartOrderProcessing(ctx context.Context) {
// for {
// 	select {
// 	case <-ctx.Done():
// 		log.Println("Stopping order processing due to context cancellation")
// 		return
// 	default:
// 		orders, err := repo.GetNewOrders(ctx)
// 		if err != nil {
// 			log.Printf("Failed to get new orders: %v", err)
// 			time.Sleep(1 * time.Minute) // Retry after some time
// 			continue
// 		}

// 		for _, order := range orders {
// 			go processOrder(ctx, repo, order)
// 		}

// 		time.Sleep(1 * time.Minute) // Polling interval to check for new orders
// 	}
// }
// }

// func processOrder(ctx context.Context, repo *repository.Repository) {
// for {
// 	select {
// 	case <-ctx.Done():
// 		log.Printf("Stopping processing for order %s due to context cancellation", order.Number)
// 		return
// 	default:
// 		accrualResponse, err := fetchAccrual(order.Number)
// 		if err != nil {
// 			log.Printf("Error fetching accrual for order %s: %v", order.Number, err)
// 			time.Sleep(1 * time.Minute) // Retry after some time
// 			continue
// 		}

// 		switch accrualResponse.Status {
// 		case "REGISTERED":
// 			err = repo.UpdateOrderStatus(ctx, order.ID, "PROCESSING")
// 		case "INVALID":
// 			err = repo.UpdateOrderStatus(ctx, order.ID, "INVALID")
// 			if err == nil {
// 				return // Stop processing
// 			}
// 		case "PROCESSING":
// 			if order.Status == "NEW" {
// 				err = repo.UpdateOrderStatus(ctx, order.ID, "PROCESSING")
// 			}
// 		case "PROCESSED":
// 			err = repo.UpdateOrderStatusAndAccrual(ctx, order.ID, "PROCESSED", accrualResponse.Accrual)
// 			if err == nil {
// 				return // Stop processing
// 			}
// 		}

// 		if err != nil {
// 			log.Printf("Error updating status for order %s: %v", order.Number, err)
// 		}

// 		time.Sleep(1 * time.Minute) // Polling interval
// 	}
// }
// }

// func fetchAccrual(orderNumber string) (*AccrualResponse, error) {
// resp, err := http.Get("http://accrual-system/api/orders/" + orderNumber)
// if err != nil {
// 	return nil, err
// }
// defer resp.Body.Close()

// var accrualResponse AccrualResponse
// if err := json.NewDecoder(resp.Body).Decode(&accrualResponse); err != nil {
// 	return nil, err
// }

// return &accrualResponse, nil
// }
