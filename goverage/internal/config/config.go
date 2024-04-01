package config

import (
	"os"

	"github.com/rs/zerolog/log"
)

type config struct {
	DBConnStr string
	APIKey    string
}

var Config *config

func LoadConfig() {
	dbConnStr := os.Getenv("GOVERAGE_DB_CONN_STR")
	if dbConnStr == "" {
		log.Fatal().Msg("GOVERAGE_DB_CONN_STR is required")
	}

	apiKey := os.Getenv("GOVERAGE_API_KEY")
	if apiKey == "" {
		log.Fatal().Msg("GOVERAGE_API_KEY is required")
	}

	Config = &config{
		DBConnStr: dbConnStr,
		APIKey:    apiKey,
	}
}
