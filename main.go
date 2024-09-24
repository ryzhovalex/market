package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

var VERSION string = "0.1.0"
var state State

type CoinAmount = int32
type Id = int64
type Time = int64
type Dict = map[string]any
type Query = Dict
type GetQuery = Query

// Length of the standard time chunk.
//
// All time-related activites (such as jobs) relate to this chunk, unless
// other is explicitly denoted. For example `primary` job will rely on this
// chunk: so if the chunk is 30 minutes, the primary job will refer to that
// value.
//
// There are other chunks besides time, but they are useless for now to denote
// in form of static vars, e.g. `AMOUNT_CHUNK = 1`.
var TIME_CHUNK = 30 * 60

func utc() Time {
	return time.Now().Unix()
}

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
	Created Time   `json:"created"`
	Cmd     string `json:"cmd"`
}

func Unwrap(err error) {
	if err != nil {
		panic(err)
	}
}

func Print(obj ...any) {
	fmt.Println(obj...)
}

func Printf(f string, obj ...any) {
	Print(fmt.Sprintf(f, obj...))
}

func writeState() {
	b, err := json.MarshalIndent(state, "", "    ")
	Unwrap(err)
	workingDir, e := getWorkingDir()
	Unwrap(e)
	err = os.WriteFile(workingDir+"/var/state.json", b, 0644)
	Unwrap(err)
}

var lastInp string = ""

func readInp() ([]string, Err) {
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

type ErrData struct {
	msg  string
	code string
}

type Err interface {
	Error() string
	Msg() string
	Code() string
}

func (e *ErrData) Error() string {
	return fmt.Sprintf("%s:: %s", e.code, e.msg)
}

func (e *ErrData) Msg() string {
	return e.msg
}

func (e *ErrData) Code() string {
	return e.code
}

func NewErr(msg string, code string) *ErrData {
	if code == "" {
		code = "err"
	}
	return &ErrData{msg, code}
}

func cliBuy(args []string) Err {
	itemKey := args[0]
	item, ok := KEY_TO_ITEM[itemKey]
	if !ok {
		return NewErr("No such item "+itemKey, "")
	}
	if item.Price > state.Balance {
		return NewErr(fmt.Sprintf(
			"Not enough coins (current=%d) to buy item %s with price %d",
			state.Balance,
			item.Key,
			item.Price,
		), "")
	}
	oldBalance := state.Balance
	state.Balance -= item.Price
	Printf(
		"Bought item %s for a price %d. Balance change: %d -> %d",
		item.Key, item.Price, oldBalance, state.Balance,
	)
	return nil
}

func cliJob(args []string) Err {
	jobKey := args[0]
	job, ok := KEY_TO_JOB[jobKey]
	if !ok {
		return NewErr("No such job "+jobKey, "")
	}
	oldBalance := state.Balance
	state.Balance += job.Reward
	Printf(
		"Completed job %s with reward %d. Balance change: %d -> %d",
		job.Key, job.Reward, oldBalance, state.Balance,
	)
	return nil
}

func quit(args []string) Err {
	writeState()
	return NewErr("Interrupt", "interrupt_err")
}

func repeat(args []string) Err {
	repeatArg, classicErr := strconv.Atoi(args[0])
	Unwrap(classicErr)

	cmd := args[1]
	f, ok := FNS[cmd]
	if !ok {
		return NewErr("Unrecognized command \""+cmd+"\"", "")
	}
	fnArgs := args[2:]
	for i := 0; i < repeatArg; i++ {
		err := f(fnArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

// Functions that are considered as transactional - after their execution a
// transaction record will be created.
var TRANSACTION_FN_KEYS = []string{"buy", "job"}
var FNS = map[string]func(args []string) Err{
	"q":       quit,
	"dir":     cliDir,
	"version": cliVersion,
	"buy":     cliBuy,
	"job":     cliJob,
	"balance": cliBalance,
	"items":   cliItems,
	"jobs":    cliJobs,
}

func cliItems(args []string) Err {
	Print("ITEMS")
	for k, v := range KEY_TO_ITEM {
		Printf("\t%s: %d coins", k, v.Price)
	}
	return nil
}

func cliJobs(args []string) Err {
	Print("JOBS")
	for k, v := range KEY_TO_JOB {
		Printf("\t%s: %d coins", k, v.Reward)
	}
	return nil
}

func cliVersion(args []string) Err {
	Printf("Version: %s", VERSION)
	return nil
}

func ToErr(e error) Err {
	return NewErr(e.Error(), "err")
}

func cliDir(args []string) Err {
	d, e := getWorkingDir()
	if e != nil {
		return e
	}
	Printf("Working dir: %s", d)
	return nil
}

func cliBalance(args []string) Err {
	Printf("Balance: %d", state.Balance)
	return nil
}

func loop() {
	Print("Welcome to Market!")
	Printf("Balance: %d", state.Balance)

	for {
		inp, err := readInp()
		if err != nil {
			Print(err)
			if err.Code() == "interrupt_err" {
				break
			}
			continue
		}
		if strings.ReplaceAll(lastInp, " ", "") == "" {
			continue
		}
		cmd := inp[0]
		args := inp[1:]

		var f func(args []string) Err
		var ok bool
		// Handle repeat under special conditions since it references `FNS`
		// under the hood, so it cannot be put as part of `FNS`.
		if cmd == "r" {
			f = repeat
		} else {
			f, ok = FNS[cmd]
			if !ok {
				Print("err:: Unrecognized command \"" + cmd + "\"")
				continue
			}
		}

		err = f(args)
		if err != nil {
			Print(err)
			if err.Code() == "interrupt_err" {
				break
			}
			continue
		}
		// Do not save transaction on error.
		if slices.Contains(TRANSACTION_FN_KEYS, cmd) {
			state.Transactions = append(
				state.Transactions,
				Transaction{utc(), lastInp},
			)
		}
	}
}

var KEY_TO_ITEM = map[string]Item{}
var KEY_TO_JOB = map[string]Job{}

func jsonLoad(path string, v any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, &v)
	if err != nil {
		return err
	}
	return nil
}

func getWorkingDir() (string, Err) {
	useCwdDir := os.Getenv("MARKET_USE_CWD")
	if useCwdDir == "1" {
		r, baserr := os.Getwd()
		if baserr != nil {
			return "", ToErr(baserr)
		}
		return r, nil
	}

	ex, err := os.Executable()
	if err != nil {
		return "", ToErr(err)
	}
	exPath := filepath.Dir(ex)
	return exPath, nil
}

func init() {
	workingDir, e := getWorkingDir()
	Unwrap(e)

	err := jsonLoad(workingDir+"/var/state.json", &state)
	Unwrap(err)

	var items []Item
	err = jsonLoad(workingDir+"/data/item.json", &items)
	for _, v := range items {
		KEY_TO_ITEM[v.Key] = v
	}

	Unwrap(err)
	var jobs []Job
	err = jsonLoad(workingDir+"/data/job.json", &jobs)
	for _, v := range jobs {
		KEY_TO_JOB[v.Key] = v
	}
	Unwrap(err)
}

func main() {
	defer writeState()
	// Setup interrupt signals so we actually can intercept keyboard interrupt
	// in things like `inp.Scan()`. Don't know entirely why it needs, but by
	// trial and error i've managed to do this.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	loop()
}
