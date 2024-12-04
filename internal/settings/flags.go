package settings

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
)

// - адрес и порт запуска сервиса: переменная окружения ОС `RUN_ADDRESS` или флаг `-a`
// - адрес подключения к базе данных: переменная окружения ОС `DATABASE_URI` или флаг `-d`
// - адрес системы расчёта начислений: переменная окружения ОС `ACCRUAL_SYSTEM_ADDRESS` или флаг `-r`

type InitedFlags struct {
	Endpoint      string
	DBSettings    string
	RecalcAddress string
}

func Parse() *InitedFlags {
	end := flag.String("a", ":8080", "endpoint address")
	flagDBSettings := flag.String("d", "", "Адрес подключения к БД")
	flagRecalcMetrics := flag.String("r", ":8010", "Адрес системы расчёта начислений")

	flag.Parse()
	_ = godotenv.Load()

	endpoint := *end
	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		endpoint = envRunAddr
	}

	recalcMetrics := *flagRecalcMetrics
	if envCalcAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envCalcAddr != "" {
		recalcMetrics = envCalcAddr
	}

	dbSettings := *flagDBSettings
	if envRunDBSettings := os.Getenv("DATABASE_URI"); envRunDBSettings != "" {
		dbSettings = envRunDBSettings
	}
	return &InitedFlags{
		Endpoint:      endpoint,
		DBSettings:    dbSettings,
		RecalcAddress: recalcMetrics,
	}

}
