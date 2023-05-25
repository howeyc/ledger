# Export

We can export transactions in CSV format.

## Example

`$ ledger -f ledger.dat export > transactions.csv`

By default columns are comma separated. To use another delimiter use the `--delimiter` flag e.g.

`$ ledger -f ledger.dat --delimiter $'\t' export > transactions.csv`

`$'\t'` will produce a literal tab character in Bash shell environment.
