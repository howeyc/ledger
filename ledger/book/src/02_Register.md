# Register

Run `ledger -f ledger.dat reg` to see all transactions in register format.
Since we aren't specifying a specific account, we will get all postings for
all transactions and the running total will sum to zero, as all transactions
balance.

`$ ledger -f ledger.dat reg`

## Payee

Let's see how much money we've spend at the "Grocery Store" each month. Also,
to keep from seeing every posting, we are going to specify that we only want to
see postings in the "Expenses" accounts. This will allow us to easily see a
running total in the last column.

Run `ledger -f ledger.dat reg --payee "Grocery Store" --period Monthly Expenses`

`$ ledger -f ledger.dat reg --payee "Grocery Store" --period Monthly Expenses`

## Accounts

Let's track down all the times we used our Credit Card.

Run `ledger -f ledger.dat reg MasterCard`

`$ ledger -f ledger.dat reg MasterCard`
