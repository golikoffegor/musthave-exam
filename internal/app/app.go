package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"musthave-exam/internal/model"
	"musthave-exam/internal/settings"
	"musthave-exam/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/sirupsen/logrus"
)

type app struct {
	Logger *logrus.Logger
	DB     *sql.DB
	Flags  *settings.InitedFlags
	Ctx    context.Context
	CncF   context.CancelFunc
}

func NewApp() (*app, error) {
	parsed := settings.Parse()
	fmt.Printf("parsed: %v\n", parsed)
	//set logger
	logg := logrus.New()
	// logg.SetFormatter(&logrus.JSONFormatter{})
	// logg.SetLevel(logrus.InfoLevel)
	logg.SetLevel(logrus.DebugLevel)
	// file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	// if err != nil {
	// 	logg.Fatal(err)
	// }
	// defer file.Close()
	// logg.SetOutput(file)
	fmt.Printf("parsed.DBSettings: %v\n", parsed.DBSettings)
	//set db connect
	db, err := NewConnection(parsed.DBSettings)
	if err != nil {
		return nil, err
	}
	goose.SetBaseFS(migrations.Migrations)

	cnt, cancel := context.WithCancel(context.Background())

	return &app{
		Logger: logg,
		DB:     db,
		Flags:  parsed,
		Ctx:    cnt,
		CncF:   cancel,
	}, nil
}

func NewConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logrus.Error("failed to create a database connection", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		logrus.Error("failed to ping the database", err)
		return nil, err
	}

	return db, err
}

func FetchAccrual(address string, transactionID string) (*model.AccrualResponse, error) {
	url := fmt.Sprintf(address+"/api/orders/%s", transactionID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("FetchAccrual resp.StatusCode: %v\n", resp.StatusCode)

	var accrualResponse model.AccrualResponse
	if resp.StatusCode == http.StatusNoContent {
		return &accrualResponse, nil
	}

	if err := json.NewDecoder(resp.Body).Decode(&accrualResponse); err != nil {
		return nil, err
	}

	return &accrualResponse, nil
}

// func NewConnection(driver, dsn string) (*sqlx.DB, error) {
// 	db, err := sqlx.Open(driver, dsn)
// 	if err != nil {
// 		slog.Error("failed to create a database connection", err)
// 		return nil, err
// 	}

// 	if err = db.Ping(); err != nil {
// 		slog.Error("failed to ping the database", err)
// 		return nil, err
// 	}

// 	return db, nil
// }
