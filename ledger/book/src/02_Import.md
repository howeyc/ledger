# Import

We can import transactions in CSV format, and product ledger transactions.
Transactions are classified using best-likely match based on payee descriptions.
Matches do not need to be exact matches, it's based on probablility determined
by learning from existing transactions. The more existing transactions in your
ledger file, the better the matches will be.

## Example

Example transactions from your credit card csv download.

`$ cat transactions.csv`

Let's run our import, making sure to specify the correct date-format to match
the CSV file.

Run `ledger -f ledger.dat --date-format "01/02/06" import MasterCard transactions.csv`

`$ ledger -f ledger.dat --date-format "01/02/06" import MasterCard transactions.csv`

These are not written to our ledger file, just displayed. Once we are satisfied
with the transactions we can write them to our ledger file by
running `ledger -f ledger.dat --date-format "01/02/06" import MasterCard transactions.csv >> ledger.dat`
