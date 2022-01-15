# Equity

Some users like to keep ledger files for each year. To aid in creating a new
starting balance for the next file, we can use the `ledger equity` command to
generate the required transaction to have the correct starting balances.

Let's start 2021, using all transactions upto the end of 2020. As the end date
on the command line is not included, we can use 2021/01/01 as the end date.

Run `ledger -f ledger.dat equity -e "2021/01/01"`

`$ ledger -f ledger.dat equity -e "2021/01/01"`
