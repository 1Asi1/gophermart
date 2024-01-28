package config

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
)

type Config struct {
	ServerAddr  string
	DBConnDSN   string
	AccrualAddr string
}

func New(log zerolog.Logger) Config {
	l := log.With().Str("config", "New").Logger()

	var cfg Config
	add := flag.String("a", "127.0.0.1:8080", "address and port to run service")
	accrualAdd := flag.String("r", "127.0.0.1:8081", "address accrual servers")
	db := flag.String("d", "", "dsn connecting to postgres")
	flag.Parse()

	addrEnv, ok := os.LookupEnv("RUN_ADDRESS")
	if ok {
		l.Info().Msgf("server address value: %s", addrEnv)
		cfg.ServerAddr = addrEnv
	} else {
		l.Info().Msgf("server address value: %s", *add)
		cfg.ServerAddr = *add
	}

	accrualAddrEnv, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if ok {
		l.Info().Msgf("accrual server address value: %s", addrEnv)
		cfg.AccrualAddr = accrualAddrEnv
	} else {
		l.Info().Msgf("accrual server address value: %s", *add)
		cfg.AccrualAddr = *accrualAdd
	}

	dbDSN, ok := os.LookupEnv("DATABASE_URI")
	if ok {
		cfg.DBConnDSN = dbDSN
	} else {
		cfg.DBConnDSN = *db
	}

	return cfg
}
