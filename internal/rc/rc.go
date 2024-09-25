package rc

import (
	"encoding/json"
	"os"

	"github.com/slimebones/market/internal/err"
)

var DEFAULT_APPRC_PATH string = "apprc.yml"
var APPRC_ENV string = "MARKET_APPRC"

func Load() err.E {
	var apprcPath string
	if apprcPath = os.Getenv(APPRC_ENV); apprcPath == "" {
		apprcPath = DEFAULT_APPRC_PATH
	}

	content, e := os.ReadFile(apprcPath)
	if e != nil {
		return err.FromBase()
	}
	e = json.Unmarshal(content, &v)
	if e != nil {
		return err.FromBase(e)
	}
	return nil
}
