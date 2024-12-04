package handler

import (
	"context"
	"encoding/json"
	"io"
	"musthave-exam/internal/model"
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("secret-key"))

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

	// Декодируем JSON из []byte в структуру User
	if err := json.Unmarshal(body, &user); err != nil {
		h.log.WithField("JSON", err.Error()).Info(model.ErrFailedToDecodeJSON.Error())
		http.Error(w, model.ErrFailedToDecodeJSON.Error()+": "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	userID, err := h.repo.RegisterUser(ctx, user)
	if err != nil {
		if err.Error() == model.ErrLoginAlreadyTaken.Error() {
			h.log.WithError(err).Warning(err.Error())
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.log.WithError(err).Error(model.ErrInternalServer.Error())
		http.Error(w, model.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	h.setAuth(w, r, userID)
	h.log.WithField("user_id", userID).Info("User registered and logged in")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
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

	// Декодируем JSON из []byte в структуру User
	if err := json.Unmarshal(body, &credentials); err != nil {
		h.log.WithField("JSON", err.Error()).Info(model.ErrFailedToDecodeJSON.Error())
		http.Error(w, model.ErrFailedToDecodeJSON.Error()+": "+err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := h.repo.LoginUser(context.Background(), credentials)
	if err != nil {
		if err.Error() == model.ErrLoginAlreadyTaken.Error() {
			h.log.WithError(err).Warning(model.ErrLoginAlreadyTaken.Error())
			http.Error(w, model.ErrLoginAlreadyTaken.Error(), http.StatusUnauthorized)
			return
		}
		h.log.WithError(err).Error(model.ErrInternalServer.Error())
		http.Error(w, model.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	h.setAuth(w, r, userID)
	h.log.WithField("user_id", userID).Info("User logged in")
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

	h.log.WithField("user_id", userID).Info("User data retrieved")
	json.NewEncoder(w).Encode(user)
}
