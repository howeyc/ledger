#!/usr/bin/env bash

./ledger --prof "bal.pprof" bal > /dev/null
./ledger --prof "reg.pprof" reg > /dev/null
./ledger --prof "print.pprof" print > /dev/null
./ledger --prof "stats.pprof" stats > /dev/null

rm default.pgo

go tool pprof -proto reg.pprof bal.pprof print.pprof stats.pprof > default.pgo

rm bal.pprof
rm reg.pprof
rm print.pprof
rm stats.pprof
