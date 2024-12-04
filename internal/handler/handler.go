package handler

import (
	"musthave-exam/internal/repository"
	"musthave-exam/internal/settings"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo repository.Repository
	// rt repository.TransactionRepository
	// logger   repository
	log *logrus.Logger
	cfg *settings.InitedFlags
}

// NewHandler создает экземпляр Handler
func NewHandler(storage repository.Repository, logger *logrus.Logger, cnf *settings.InitedFlags) *Handler {
	return &Handler{
		repo: storage,
		log:  logger,
		cfg:  cnf,
	}
}
