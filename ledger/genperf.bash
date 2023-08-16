#!/usr/bin/env bash

cbench --export-markdown perf-stats.md "ledger stats" "./ledger stats" "hledger stats"
cbench --export-markdown perf-bal.md "ledger bal" "./ledger bal" "hledger bal"
cbench --export-markdown perf-reg.md "ledger reg" "./ledger reg" "hledger reg"
cbench --export-markdown perf-print.md "ledger print" "./ledger print" "hledger print"

