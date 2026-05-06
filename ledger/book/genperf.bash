#!/usr/bin/env bash

cbench --export-markdown perf-stats.md "ledger stats" "../ledger stats" "hledger stats" "rledger report stats"
cbench --export-markdown perf-bal.md "ledger bal" "../ledger bal" "hledger bal" "rledger report balances"
cbench --export-markdown perf-reg.md "ledger reg" "../ledger reg" "hledger reg" "rledger report register"
cbench --export-markdown perf-print.md "ledger print" "../ledger print" "hledger print" "rledger format"

echo "# Performance" > perf.md
echo "" >> perf.md
echo "Comparison between various ledger-like applications:" >> perf.md
echo "" >> perf.md
echo "- ledger-go" >> perf.md
echo "- [ledger-cli](https://ledger-cli.org)" >> perf.md
echo "- [hledger](https://hledger.org)" >> perf.md
echo "- [rledger](https://rustledger.github.io)" >> perf.md
echo "" >> perf.md

echo "## Stats" >> perf.md
echo "" >> perf.md
cat perf-stats.md | sed -e 's/\.\.\/ledger/ledger-go/g' | sed -e 's/ledger /ledger-cli /g' | sed -e 's/hledger-cli/hledger/g' |  sed -e 's/rledger-cli/rledger/g'  >> perf.md
echo "" >> perf.md

echo "## Balance" >> perf.md
echo "" >> perf.md
cat perf-bal.md | sed -e 's/\.\.\/ledger/ledger-go/g' | sed -e 's/ledger /ledger-cli /g' | sed -e 's/hledger-cli/hledger/g' |  sed -e 's/rledger-cli/rledger/g' >> perf.md
echo "" >> perf.md

echo "## Register" >> perf.md
echo "" >> perf.md
cat perf-reg.md | sed -e 's/\.\.\/ledger/ledger-go/g' | sed -e 's/ledger /ledger-cli /g' | sed -e 's/hledger-cli/hledger/g' |  sed -e 's/rledger-cli/rledger/g' >> perf.md
echo "" >> perf.md

echo "## Print" >> perf.md
echo "" >> perf.md
cat perf-print.md | sed -e 's/\.\.\/ledger/ledger-go/g' | sed -e 's/ledger /ledger-cli /g' | sed -e 's/hledger-cli/hledger/g' |  sed -e 's/rledger-cli/rledger/g' >> perf.md
echo "" >> perf.md

rm perf-stats.md perf-bal.md perf-reg.md perf-print.md
mv perf.md src/Performance.md
