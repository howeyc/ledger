#!/usr/bin/env bash

./ledger --profile "bal.pprof" bal > /dev/null
./ledger --profile "reg.pprof" reg > /dev/null
./ledger --profile "print.pprof" print > /dev/null

rm default.pgo

go tool pprof -proto reg.pprof bal.pprof print.pprof > default.pgo

rm bal.pprof
rm reg.pprof
rm print.pprof
