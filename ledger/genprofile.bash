#!/usr/bin/env bash

for i in $(seq 1 10);
do
	./ledger --prof "bal$i.pprof" bal > /dev/null
	./ledger --prof "reg$i.pprof" reg > /dev/null
	./ledger --prof "print$i.pprof" print > /dev/null
	./ledger --prof "stats$i.pprof" stats > /dev/null
done

rm default.pgo

go tool pprof -proto reg{1..10}.pprof bal{1..10}.pprof print{1..10}.pprof stats{1..10}.pprof > default.pgo

rm bal{1..10}.pprof
rm reg{1..10}.pprof
rm print{1..10}.pprof
rm stats{1..10}.pprof
