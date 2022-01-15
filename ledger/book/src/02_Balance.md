# Balance

Run `ledger -f ledger.dat bal` to see a balance report.

`$ ledger -f ledger.dat bal`

## Net Worth

You can show specific accounts by applying a filter, which is case senstive.
For example, let's get our net worth,
run `ledger -f ledger.dat bal Assets Liabilities`

`$ ledger -f ledger.dat bal Assets Liabilities`

## By Period

We can see our balances segmented by a time period. For example, let's see all
our expenses for each month,
run `ledger -f ledger.dat --period Monthly bal Expenses`

`$ ledger -f ledger.dat --period Monthly bal Expenses`

## Account Depth

That's a lot of accounts, let's trim it down to see it summed up to the second
level. Run `ledger -f ledger.dat --period Monthly --depth 2 bal Expenses`

`$ ledger -f ledger.dat --period Monthly --depth 2 bal Expenses`
