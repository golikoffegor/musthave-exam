package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/golikoffegor/musthave-exam/internal/model"
)

// var store = sessions.NewCookieStore([]byte("secret-key"))

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user model.User
	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, model.ErrErrorRequestBody.Error()+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем тело запроса на пустоту
	if len(body) == 0 {
		h.log.Info(model.ErrEmptyRequestBody.Error())
		http.Error(w, model.ErrEmptyRequestBody.Error(), http.StatusBadRequest)
		return
	}

	if err := isValidJSON(body); err != nil {
		h.log.WithField("ValidJSON", err.Error()).Info(model.ErrFailedToDecodeJSON.Error())
		http.Error(w, model.ErrFailedToDecodeJSON.Error()+": "+err.Error(), http.StatusBadRequest)
		return
	}

	// Декодируем JSON из []byte в структуру User
	if err := json.Unmarshal(body, &user); err != nil {
		h.log.WithField("JSON", err.Error()).Info(model.ErrFailedToDecodeJSON.Error())
		http.Error(w, model.ErrFailedToDecodeJSON.Error()+": "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	if len(user.Login) == 0 ||
		len(user.Password) == 0 {
		h.log.WithError(err).Error(model.ErrInternalServer.Error())
		http.Error(w, model.ErrInternalServer.Error(), http.StatusBadRequest)
	}

	userID, err := h.repo.RegisterUser(ctx, user)
	if err != nil {
		fmt.Printf("userID: %v\n", userID)
		if err.Error() == model.ErrLoginAlreadyTaken.Error() {
			h.log.WithError(err).Warning(err.Error())
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.log.WithError(err).Error(model.ErrInternalServer.Error())
		http.Error(w, model.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	err = h.setAuth(w, r, userID)
	if err != nil {
		h.log.WithError(err).Info("h.setAuth failed")
	}

	h.log.
		WithField("userID", userID).
		Info("RegisterHandler")

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var credentials model.User
	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, model.ErrErrorRequestBody.Error()+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем тело запроса на пустоту
	if len(body) == 0 {
		h.log.Info(model.ErrEmptyRequestBody.Error())
		http.Error(w, model.ErrEmptyRequestBody.Error(), http.StatusBadRequest)
		return
	}

	if err := isValidJSON(body); err != nil {
		h.log.WithField("ValidJSON", err.Error()).Info(model.ErrFailedToDecodeJSON.Error())
		http.Error(w, model.ErrFailedToDecodeJSON.Error()+": "+err.Error(), http.StatusBadRequest)
		return
	}

	var raw map[string]json.RawMessage
	err = json.Unmarshal(body, &raw)
	if err != nil {
		h.log.WithError(err).Warning(model.ErrInvalidLoginAndPass.Error())
		http.Error(w, model.ErrInvalidLoginAndPass.Error(), http.StatusUnauthorized)
		return
	}

	if len(raw) > 2 { // Here you can check for extra fields
		h.log.WithField("JSON", model.ErrInvalidLoginAndPass.Error()).Info(model.ErrFailedToDecodeJSON.Error())
		http.Error(w, model.ErrInvalidLoginAndPass.Error(), http.StatusBadRequest)
		return
	}

	// Декодируем JSON из []byte в структуру User
	if err := json.Unmarshal(body, &credentials); err != nil {
		h.log.WithField("JSON", err.Error()).Info(model.ErrFailedToDecodeJSON.Error())
		http.Error(w, model.ErrFailedToDecodeJSON.Error()+": "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(credentials.Login) == 0 ||
		len(credentials.Password) == 0 {
		h.log.WithError(err).Error(model.ErrInternalServer.Error())
		http.Error(w, model.ErrInternalServer.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := h.repo.LoginUser(context.Background(), credentials)
	if err != nil {
		if err.Error() == model.ErrInvalidLoginAndPass.Error() {
			h.log.WithError(err).Warning(model.ErrInvalidLoginAndPass.Error())
			http.Error(w, model.ErrInvalidLoginAndPass.Error(), http.StatusUnauthorized)
			return
		}
		h.log.WithError(err).Error(model.ErrInternalServer.Error())
		http.Error(w, model.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	err = h.setAuth(w, r, userID)
	if err != nil {
		h.log.WithError(err).Info("h.setAuth failed")
	}

	h.log.
		WithField("userID", userID).
		Info("LoginUserHandler")

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.isValidAuth(w, r)
	if !ok {
		http.Error(w, model.ErrNotAuthorized.Error(), http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUser(context.Background(), userID)
	if err != nil {
		if err.Error() == model.ErrUserNotFound.Error() {
			h.log.WithError(err).Warning(model.ErrUserNotFound.Error())
			http.Error(w, model.ErrUserNotFound.Error(), http.StatusNotFound)
			return
		}
		h.log.WithError(err).Error("Failed to retrieve user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.log.
		WithField("userID", userID).
		Debug("GetUserHandler")

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		h.log.WithError(err).Info("json.NewEncoder failed")
	}
}
