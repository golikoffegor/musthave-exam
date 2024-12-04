package handler

import (
	"encoding/json"
	"io"
	"musthave-exam/internal/model"
	"musthave-exam/internal/repository"
	"net/http"
	"strings"
)

func (h *Handler) AddOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.isValidAuth(w, r)
	if !ok {
		http.Error(w, model.ErrNotAuthorized.Error(), http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, model.ErrErrorRequestBody.Error(), http.StatusInternalServerError)
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if !h.isValidOrderNumber(orderNumber) {
		http.Error(w, model.ErrInvalidOrderNumber.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.log.WithField("orderNumber", orderNumber).Debug("AddOrder", userID)
	ctx := r.Context()
	transaction, err := h.repo.AddOrder(ctx, userID, orderNumber)
	if err != nil {
		if err.Error() == model.ErrAddExistsOrder.Error() {
			h.log.WithField("order already added by other user", orderNumber).Debug("AddOrder", userID)
			http.Error(w, orderNumber, http.StatusConflict)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if transaction != nil && len((*transaction).ID) > 0 {
		h.log.WithField("order added", orderNumber).Debug("AddOrder", userID)
		w.WriteHeader(http.StatusAccepted)
	} else {
		h.log.WithField("order already added by cur user", orderNumber).Debug("AddOrder", userID)
		w.WriteHeader(http.StatusOK)
	}

	w.Write([]byte(orderNumber))
}

func (h *Handler) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.isValidAuth(w, r)
	if !ok {
		http.Error(w, model.ErrNotAuthorized.Error(), http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	orders, err := h.repo.GetOrders(ctx, userID, repository.Debit)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.log.WithField("len orders", len(orders)).Debug("GetOrders", userID)
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	h.log.
		WithField("userID", userID).
		WithField("orders", orders).
		Debug("GetOrdersHandler")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *Handler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.isValidAuth(w, r)
	if !ok {
		http.Error(w, model.ErrNotAuthorized.Error(), http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	balance, err := h.repo.GetBalance(ctx, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.log.WithField("balance", balance).Debug("GetBalance", userID)
	response, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *Handler) GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.isValidAuth(w, r)
	if !ok {
		http.Error(w, model.ErrNotAuthorized.Error(), http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	orders, err := h.repo.GetOrders(ctx, userID, repository.Withdraw)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var withdrawals []model.Withdraw
	for _, order := range orders {
		withdrawals = append(withdrawals, model.Withdraw{
			ID:   order.ID,
			Summ: order.Summ,
			Date: order.Date,
		})
	}
	response, err := json.Marshal(withdrawals)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *Handler) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.isValidAuth(w, r)
	if !ok {
		http.Error(w, model.ErrNotAuthorized.Error(), http.StatusUnauthorized)
		return
	}

	var withdrawRequest model.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&withdrawRequest); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if !h.isValidOrderNumber(withdrawRequest.Order) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	err := h.repo.Withdraw(ctx, userID, withdrawRequest.Order, withdrawRequest.Sum)
	if err != nil {
		if err.Error() == model.ErrIncFunds.Error() {
			http.Error(w, model.ErrIncFunds.Error(), http.StatusPaymentRequired)
			return
		} else {
			h.log.WithError(err).WithField("WithdrawHandler", withdrawRequest)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
