package apprc

import (
	"encoding/json"
	"os"

	"github.com/slimebones/market/internal/dict"
	"github.com/slimebones/market/internal/err"
)

var DEFAULT_APPRC_PATH string = "apprc.yml"
var APPRC_ENV string = "MARKET_APPRC"
var rc dict.Dict

func Load() err.E {
	var apprcPath string
	if apprcPath = os.Getenv(APPRC_ENV); apprcPath == "" {
		apprcPath = DEFAULT_APPRC_PATH
	}

	content, be := os.ReadFile(apprcPath)
	if be != nil {
		return err.FromBase(be)
	}
	be = json.Unmarshal(content, &rc)
	if be != nil {
		return err.FromBase(be)
	}
	return nil
}
