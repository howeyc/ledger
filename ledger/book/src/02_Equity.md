# Equity

Some users like to keep ledger files for each year. To aid in creating a new
starting balance for the next file, we can use the `ledger equity` command to
generate the required transaction to have the correct starting balances.

Let's start 2022, using all transactions up to the end of 2021. As the end date
on the command line is not included, we can use 2022/01/01 as the end date.

Run `ledger -f ledger.dat equity -e "2022/01/01"`

`$ ledger -f ledger.dat equity -e "2022/01/01"`
