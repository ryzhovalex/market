package app

import (
	"os"

	"github.com/slimebones/market/internal/apprc"
	"github.com/slimebones/market/internal/err"
)

var MODE_ENV string = "MARKET_MODE"
var MODE string

func Init() err.E {
	MODE = os.Getenv(MODE_ENV)
	if MODE == "" {
		panic("Please set " + MODE_ENV)
	}

	e := apprc.Load(MODE)
	err.Unwrap(e)

	return nil
}
