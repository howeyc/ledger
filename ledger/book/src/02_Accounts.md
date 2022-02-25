# Accounts

Run `ledger -f ledger.dat accounts` to see an account list.

`$ ledger -f ledger.dat accounts`

## Only Leaf (Max Depth) Accounts

If we are only intersted in the highest depth accounts and not interested
in seeing all the parent account levels we can get that, just
run `ledger -f ledger.dat accounts -l`

`$ ledger -f ledger.dat accounts -l`

## Matching Depth Accounts

This is mostly useful for autocomplete functions. You can use this to get
accounts matching a filter, and at the same depth as the filter.

For instance, let's get all Assets accounts by running
`ledger -f ledger.dat accounts -m Assets:`

`$ ledger -f ledger.dat accounts -m Assets:`

