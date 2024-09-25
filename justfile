set shell := ["nu", "-c"]
set dotenv-load

lint:
    @ go fmt

test t="":
    @ if "{{t}}" == "" { go test } else { go test -run {{t}} }

check: lint test

run:
    @ go run .

install:
    @ go build
    @ mv market.exe ~/market/bin/release/market
