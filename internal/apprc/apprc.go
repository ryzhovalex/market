package apprc

import (
	"os"

	"github.com/slimebones/market/internal/dict"
	"github.com/slimebones/market/internal/err"
	"gopkg.in/yaml.v3"
)

var DEFAULT_APPRC_PATH string = "apprc.yml"
var APPRC_ENV string = "MARKET_APPRC"
var rc dict.Dict

func Load(build string, mode string) err.E {
	var apprcPath string
	if apprcPath = os.Getenv(APPRC_ENV); apprcPath == "" {
		apprcPath = DEFAULT_APPRC_PATH
	}

	content, be := os.ReadFile(apprcPath)
	if be != nil {
		return err.FromBase(be)
	}
	be = yaml.Unmarshal(content, &rc)
	if be != nil {
		return err.FromBase(be)
	}
	return nil
}

func Get(key string) (dict.Dict, err.E) {
	r, ok := rc[key]
	if !ok {
		return nil, err.New(
			"No such configuration with key "+key,
			err.CODE_NOT_FOUND)
	}
	rDict, ok := r.(dict.Dict)
	return rDict, nil
}
