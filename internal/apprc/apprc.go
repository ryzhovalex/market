package apprc

import (
	"fmt"
	"os"
	"regexp"

	"github.com/slimebones/market/internal/dict"
	"github.com/slimebones/market/internal/errors"
	"github.com/slimebones/market/internal/paths"
	"gopkg.in/yaml.v3"
)

var DEFAULT_APPRC_PATH string = "apprc.yml"
var APPRC_ENV string = "MARKET_APPRC"
var rc dict.Dict
var rcVars = dict.Dict{}
var varnameToResolver = map[string]Resolver{}
var VAR_REGEX = regexp.MustCompile(`\$\{([a-z0-9_\.]+)\}`)

type Resolver struct {
	fn   func(v string)
	deps []string
}

func Load(mode string) errors.E {
	var apprcPath string
	if apprcPath = os.Getenv(APPRC_ENV); apprcPath == "" {
		apprcPath = DEFAULT_APPRC_PATH
	}

	content, be := os.ReadFile(apprcPath)
	if be != nil {
		return errors.FromBase(be)
	}
	be = yaml.Unmarshal(content, &rc)
	if be != nil {
		return errors.FromBase(be)
	}

	e := setDefaultRcVars(mode)
	if e != nil {
		return e
	}

	e = collect(mode)
	if e != nil {
		return e
	}

	e = link("this", rc)
	if e != nil {
		return e
	}

	e = compile(rc)
	if e != nil {
		return e
	}

	return nil
}

func setDefaultRcVars(mode string) errors.E {
	rcVars["cwd"] = paths.MustGetCwd()
	rcVars["mode"] = mode
	rcVars["exe_dir"] = paths.MustGetExeDir()
	return nil
}

// Replace all variables with actual values
func compile(entry dict.Dict) errors.E {
	for varname, resolver := range varnameToResolver {
		resolved := dict.Dict{}
		for _, dep := range resolver.deps {
			depVal, ok := rcVars[dep]
			if !ok {
				return errors.New(
					"Unsatisfied dependency "+dep+" for var "+varname,
					"",
				)
			}
			resolved[dep] = depVal
		}
		VAR_REGEX.ReplaceAllString()
	}
	return nil
}

// Traverse all values first, collect their keys and variables to resolve them
// later.
//
// Evaluates literals.
func link(key string, entry dict.Dict) errors.E {
	for k, v := range entry {
		varname := key + "." + k

		vDict, ok := v.(dict.Dict)
		if ok {
			e := link(varname, vDict)
			return e
		}

		vStr, ok := v.(string)
		if !ok {
			// Only strings can contain `${}` blocks, i.e. be resolvable
			rcVars[varname] = v
			continue
		}
		matches := VAR_REGEX.FindAllStringSubmatch(vStr, -1)
		if matches == nil {
			// No vars => resolve immediately
			rcVars[varname] = v
			continue
		}
		deps := convertMatchesToDeps(matches)

		varnameToResolver[varname] = Resolver{
			func(resolved string) {
				entry[k] = resolved
			},
			deps,
		}
	}
	return nil
}

func convertMatchesToDeps(matches [][]string) []string {
	// For now we assume we have [[raw_1 group_1], [raw_2 group_2], ...]
	// structure, but we don't know whether it could be different for our
	// regex.
	deps := []string{}
	for _, match := range matches {
		deps = append(deps, match[1])
	}
	return deps
}

// Collect apprc for current build and mode.
func collect(mode string) errors.E {
	default_, ok := rc["_default"]
	if !ok {
		panic("Must have `_default` mode defined")
	}
	defaultDict := mustValidateDict("default_", default_)

	// TODO: Implement arrow `->` inheritance of modes.
	modeCfgPack, ok := rc[mode]
	if modeCfgPack == nil {
		modeCfgPack = dict.Dict{}
	}
	modeCfgPackDict := mustValidateDict(mode, modeCfgPack)
	// Can overwrite since we have everything we need
	rc = defaultDict
	if ok {
		for k, v := range modeCfgPackDict {
			rc[k] = v
		}
	}
	return nil
}

func mustValidateDict(k string, v any) dict.Dict {
	d, ok := v.(dict.Dict)
	if !ok {
		panic(fmt.Sprintf("Cannot convert key %s to dict", k))
	}
	return d
}

func Get(key string) (dict.Dict, errors.E) {
	r, ok := rc[key]
	if !ok {
		return nil, errors.New(
			"No such configuration with key "+key,
			err.CODE_NOT_FOUND)
	}
	rDict, ok := r.(dict.Dict)
	return rDict, nil
}
