package repository

import (
	"context"
	"errors"
	"musthave-exam/internal/model"
	"time"
)

func (r *repo) AddOrder(ctx context.Context, userID int64, number string) (*model.Transaction, error) {
	query := `SELECT id, user_id FROM transactions WHERE id = $1 LIMIT 1`
	rows, err := r.db.QueryContext(ctx, query, number)
	if err != nil {
		r.log.WithError(err).Error("Failed to get orders")
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var order model.Transaction
		if err := rows.Scan(&order.ID, &order.UserID); err != nil {
			r.log.WithError(err).Error("Failed to scan order in add order")
			return nil, err
		}
		if order.UserID != userID {
			// r.log.Info("orders found for the given user and number")
			return nil, model.ErrAddExistsOrder
		}
	} else {
		query = `INSERT INTO transactions (id, user_id, summ, date, status, action) 
		VALUES ($1, $2, $3, $4, 'NEW', 'Debit')
		RETURNING id, user_id, summ, date, status, action`
		row := r.db.QueryRowContext(ctx, query, number, userID, 0, time.Now())
		// if err != nil {
		// 	r.log.WithError(err).Error("Failed to add order")
		// 	return nil, err
		// }
		var transaction model.Transaction
		if err := row.Scan(&transaction.ID, &transaction.UserID, &transaction.Summ, &transaction.Date, &transaction.Status, &transaction.Action); err != nil {
			r.log.WithError(err).Error(model.ErrAddOrder)
			return nil, err
		}
		r.log.WithField("transaction_id", transaction.ID).Debug("Order added successfully")
		return &transaction, nil
	}

	// Проверяем ошибку после Next()
	if err := rows.Err(); err != nil {
		r.log.WithError(err).Error("Error while iterating rows")
		return nil, err
	}

	return nil, nil
}

func (r *repo) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.log.WithError(err).Error("Failed to begin transaction")
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	query := `SELECT balance FROM "user" WHERE id = $1`
	var balance float64
	err = tx.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		r.log.WithError(err).Error("Withdraw: Failed to get balance")
		return err
	}

	if balance < sum {
		err = errors.New("insufficient funds")
		r.log.WithError(err).Warning("Insufficient funds")
		return err
	}

	updateBalance := `UPDATE "user" SET balance = balance - $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateBalance, sum, userID)
	if err != nil {
		r.log.WithError(err).Error("Failed to update balance")
		return err
	}

	insertTransaction := `INSERT INTO transactions (user_id, summ, date, status, action) VALUES ($1, $2, $3, 'NEW', 'Withdraw')`
	_, err = tx.ExecContext(ctx, insertTransaction, userID, sum, time.Now())
	if err != nil {
		r.log.WithError(err).Error("Failed to insert transaction")
		return err
	}

	return nil
}

func (r *repo) GetOrders(ctx context.Context, userID int64, act string) ([]model.Transaction, error) {
	query := `SELECT id, status, action, date, summ FROM transactions WHERE user_id = $1 AND action = $2 ORDER BY date`
	rows, err := r.db.QueryContext(ctx, query, userID, act)
	if err != nil {
		r.log.WithError(err).Error("Failed to get orders")
		return nil, err
	}
	defer rows.Close()

	var orders []model.Transaction
	for rows.Next() {
		var order model.Transaction
		if err := rows.Scan(&order.ID, &order.Status, &order.Action, &order.Date, &order.Summ); err != nil {
			r.log.WithError(err).Error("Failed to scan order")
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		r.log.WithError(err).Error("Error while iterating rows")
		return nil, err
	}

	return orders, nil
}

func (r *repo) UpdateTransactionStatus(ctx context.Context, transactionID string, status string) error {
	query := `UPDATE transactions SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, transactionID)
	return err
}

func (r *repo) UpdateTransactionStatusAndAccrual(ctx context.Context, transactionID string, status string, accrual float64) error {
	query := `UPDATE transactions SET status = $1, summ = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, accrual, transactionID)
	return err
}

func (r *repo) GetNewTransactions(ctx context.Context) ([]model.Transaction, error) {
	query := `SELECT id, user_id, summ, date, status, action FROM transactions WHERE action = 'Debit' AND status = 'NEW'`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var transaction model.Transaction
		if err := rows.Scan(&transaction.ID, &transaction.UserID, &transaction.Summ, &transaction.Date, &transaction.Status, &transaction.Action); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// func (r *repo) ListenForNewOrders(ctx context.Context) {
// 	// Подписываемся на уведомления
// 	if _, err := r.db.ExecContext(ctx, "LISTEN new_order"); err != nil {
// 		log.Fatalf("Failed to listen for new orders: %v", err)
// 	}

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Println("Stopping listening for new orders due to context cancellation")
// 			return
// 		default:
// 			// Ожидаем уведомления
// 			if notification, err := r.db.WaitForNotification(ctx); err == nil {
// 				orderID := notification.Payload
// 				order, err := repo.GetOrderByID(ctx, orderID)
// 				if err != nil {
// 					log.Printf("Failed to get order by ID %s: %v", orderID, err)
// 					continue
// 				}

// 				go processOrder(ctx, repo, order)
// 			} else {
// 				log.Printf("Failed to wait for notification: %v", err)
// 				time.Sleep(1 * time.Minute) // Retry after some time
// 			}
// 		}
// 	}
// }
