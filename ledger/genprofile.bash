#!/usr/bin/env bash

./ledger --profile "bal.pprof" bal
./ledger --profile "reg.pprof" reg
./ledger --profile "print.pprof" print

rm default.pgo

go tool pprof -proto reg.pprof bal.pprof print.pprof > default.pgo

rm bal.pprof
rm reg.pprof
rm print.pprof
