# Print

You can print your ledger file in a consistent format. Useful if you want all
transactions to be in a consistent format and your file to always be ordered by
date.

Run `ledger -f ledger.dat print`

`$ ledger -f ledger.dat print`

You can also use this if your splitting off transactions into separate files by
date range, or account.

All 2020 transactions for example `ledger -f ledger.dat -b "2020/01/01" -e "2021/01/01" print`

`$ ledger -f ledger.dat -b "2020/01/01" -e "2021/01/01" print`
