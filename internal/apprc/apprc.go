package apprc

import (
	"fmt"
	"os"

	"github.com/slimebones/market/internal/dict"
	"github.com/slimebones/market/internal/err"
	"github.com/slimebones/market/internal/log"
	"gopkg.in/yaml.v3"
)

var DEFAULT_APPRC_PATH string = "apprc.yml"
var APPRC_ENV string = "MARKET_APPRC"
var rc dict.Dict

func Load(mode string) err.E {
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

	e := compose(mode)
	if e != nil {
		return e
	}
	log.Debug(rc)

	return nil
}

// Collect apprc for current build and mode.
func compose(mode string) err.E {
	default_, ok := rc["_default"]
	if !ok {
		panic("Must have `_default` mode defined")
	}
	defaultDict := validateDictOrPanic("default_", default_)

	// TODO: Implement arrow `->` inheritance of modes.
	modeCfgPack, ok := rc[mode]
	modeCfgPackDict := validateDictOrPanic(mode, modeCfgPack)
	// Can overwrite since we have everything we need
	rc = defaultDict
	if ok {
		for k, v := range modeCfgPackDict {
			rc[k] = v
		}
	}
	return nil
}

func validateDictOrPanic(k string, v any) dict.Dict {
	d, ok := v.(dict.Dict)
	if !ok {
		panic(fmt.Sprintf("Cannot convert key %s to dict", k))
	}
	return d
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
