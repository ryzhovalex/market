package app

import (
	"os"

	"github.com/slimebones/market/internal/apprc"
	"github.com/slimebones/market/internal/err"
)

var BUILD_ENV string = "MARKET_BUILD"
var MODE_ENV string = "MARKET_MODE"
var BUILD string
var MODE string

func Init() err.E {
	BUILD = os.Getenv(BUILD_ENV)
	if BUILD != "debug" && BUILD != "release" {
		return err.New("Invalid build mode: "+BUILD, "")
	}
	MODE = os.Getenv(MODE_ENV)

	e := apprc.Load(BUILD, MODE)
	err.Unwrap(e)

	return nil
}
