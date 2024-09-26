package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"

	"github.com/slimebones/market/internal/app"
	"github.com/slimebones/market/internal/errs"
	"github.com/slimebones/market/internal/log"
	"github.com/slimebones/market/internal/paths"
	"github.com/slimebones/market/internal/times"
)

var VERSION string = "0.1.0"
var state State

type CoinAmount = int32
type Id = int64

type Item struct {
	Key   string     `json:"key"`
	Price CoinAmount `json:"price"`
}

type Job struct {
	Key    string     `json:"key"`
	Reward CoinAmount `json:"reward"`
}

type State struct {
	Balance      CoinAmount    `json:"balance"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Created times.Time `json:"created"`
	Cmd     string     `json:"cmd"`
	Version string     `json:"version"`
}

func cliWriteState(args []string) errs.Err {
	b, be := json.MarshalIndent(state, "", "    ")
	if be != nil {
		return errs.FromBase(be)
	}

	workingDir := paths.MustGetCwd()

	be = os.WriteFile(workingDir+"/var/state.json", b, 0644)
	if be != nil {
		return errs.FromBase(be)
	}

	log.Info("State written")
	return nil
}

var lastInp string = ""

func readInp() ([]string, errs.Err) {
	inp := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	ok := inp.Scan()
	if !ok {
		fmt.Print("\n")
		return nil, quit(nil)
	}

	text := inp.Text()
	lastInp = text
	return strings.Fields(text), nil
}

func cliBuy(args []string) errs.Err {
	itemKey := args[0]
	item, ok := KEY_TO_ITEM[itemKey]
	if !ok {
		return errs.New("No such item "+itemKey, "")
	}
	if item.Price > state.Balance {
		return errs.New(fmt.Sprintf(
			"Not enough coins (current=%d) to buy item %s with price %d",
			state.Balance,
			item.Key,
			item.Price,
		), "")
	}
	oldBalance := state.Balance
	state.Balance -= item.Price
	log.Infof(
		"Bought item %s for a price %d. Balance change: %d -> %d",
		item.Key, item.Price, oldBalance, state.Balance,
	)
	return nil
}

func cliJob(args []string) errs.Err {
	jobKey := args[0]
	job, ok := KEY_TO_JOB[jobKey]
	if !ok {
		return errs.New("No such job "+jobKey, "")
	}
	oldBalance := state.Balance
	state.Balance += job.Reward
	log.Infof(
		"Completed job %s with reward %d. Balance change: %d -> %d",
		job.Key, job.Reward, oldBalance, state.Balance,
	)
	return nil
}

func quit(args []string) errs.Err {
	return errs.New("Interrupt", "interrupt_err")
}

func cliRepeat(args []string) errs.Err {
	repeatArg, be := strconv.Atoi(args[0])
	errs.Unwrap(be)

	cmd := args[1]
	if !slices.Contains(TRANSACTION_FN_KEYS, cmd) {
		return errs.New(
			"Repeat function can work only with transaction keys",
			"",
		)
	}
	f, ok := FNS[cmd]
	if !ok {
		return errs.New("Unrecognized command \""+cmd+"\"", "")
	}
	fnArgs := args[2:]
	for i := 0; i < repeatArg; i++ {
		e := f(fnArgs)
		if e != nil {
			return e
		}
	}
	return nil
}

// Functions that are considered as transactional - after their execution a
// transaction record will be created.
var TRANSACTION_FN_KEYS = []string{"r", "buy", "job"}
var FNS = map[string]func(args []string) errs.Err{
	"q":       quit,
	"dir":     cliDir,
	"version": cliVersion,
	"buy":     cliBuy,
	"job":     cliJob,
	"balance": cliBalance,
	"items":   cliItems,
	"jobs":    cliJobs,
	"w":       cliWriteState,
}

func cliItems(args []string) errs.Err {
	log.Info("ITEMS")
	for k, v := range KEY_TO_ITEM {
		log.Infof("\t%s: %d coins", k, v.Price)
	}
	return nil
}

func cliJobs(args []string) errs.Err {
	log.Info("JOBS")
	for k, v := range KEY_TO_JOB {
		log.Infof("\t%s: %d coins", k, v.Reward)
	}
	return nil
}

func cliVersion(args []string) errs.Err {
	log.Infof("Version: %s", VERSION)
	return nil
}

func cliDir(args []string) errs.Err {
	d := paths.MustGetCwd()
	log.Infof("Working dir: %s", d)
	return nil
}

func cliBalance(args []string) errs.Err {
	log.Infof("Balance: %d", state.Balance)
	return nil
}

func loop() {
	log.Info("Welcome to Market!")
	log.Infof("Balance: %d", state.Balance)

	for {
		inp, e := readInp()
		if e != nil {
			if e.Code() == "interrupt_err" {
				break
			}
			log.Info(e)
			continue
		}
		if strings.ReplaceAll(lastInp, " ", "") == "" {
			continue
		}
		cmd := inp[0]
		args := inp[1:]

		var f func(args []string) errs.Err
		var ok bool
		// Handle repeat under special conditions since it references `FNS`
		// under the hood, so it cannot be put as part of `FNS`.
		if cmd == "r" {
			f = cliRepeat
		} else {
			f, ok = FNS[cmd]
			if !ok {
				log.Info("err:: Unrecognized command \"" + cmd + "\"")
				continue
			}
		}

		e = f(args)
		if e != nil {
			if e.Code() == "interrupt_err" {
				break
			}
			log.Info(e)
			continue
		}
		// Do not save transaction on error.
		if slices.Contains(TRANSACTION_FN_KEYS, cmd) {
			state.Transactions = append(
				state.Transactions,
				Transaction{times.Utc(), lastInp, VERSION},
			)
		}
	}
}

var KEY_TO_ITEM = map[string]Item{}
var KEY_TO_JOB = map[string]Job{}

func jsonLoad(path string, v any) errs.Err {
	content, be := os.ReadFile(path)
	if be != nil {
		return errs.FromBase(be)
	}
	be = json.Unmarshal(content, &v)
	if be != nil {
		return errs.FromBase(be)
	}
	return nil
}

func init() {
	e := app.Init()
	errs.Unwrap(e)

	workingDir := paths.MustGetCwd()

	e = jsonLoad(workingDir+"/var/state.json", &state)
	errs.Unwrap(e)

	var items []Item
	e = jsonLoad(workingDir+"/data/item.json", &items)
	for _, v := range items {
		KEY_TO_ITEM[v.Key] = v
	}

	errs.Unwrap(e)
	var jobs []Job
	e = jsonLoad(workingDir+"/data/job.json", &jobs)
	for _, v := range jobs {
		KEY_TO_JOB[v.Key] = v
	}
	errs.Unwrap(e)
}

func main() {
	// Setup interrupt signals so we actually can intercept keyboard interrupt
	// in things like `inp.Scan()`. Don't know entirely why it needs, but by
	// trial and error i've managed to do this.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	loop()
}
