# Export

We can export transactions in CSV format.

## Example

Run `ledger -f ledger.dat export`

`$ ledger -f ledger.dat export`

By default columns are comma separated. To use another delimiter use the `--delimiter` flag e.g.

`ledger -f ledger.dat --delimiter $'\t' export`

`$'\t'` will produce a literal tab character in Bash shell environment.

`$ ledger -f ledger.dat --delimiter '	' export`
