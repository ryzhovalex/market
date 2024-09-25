package app

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/slimebones/market/internal/apprc"
	"github.com/slimebones/market/internal/errors"
)

var MODE_ENV string = "MARKET_MODE"
var MODE string

func Init() errors.E {
	be := godotenv.Load()
	if be != nil {
		panic("Error loading .env file")
	}

	MODE = os.Getenv(MODE_ENV)
	if MODE == "" {
		panic("Please set " + MODE_ENV)
	}

	e := apprc.Load(MODE)
	errors.Unwrap(e)

	return nil
}
