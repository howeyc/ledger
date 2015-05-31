# Ledger in Go

This is a project to parse and import transactions in a ledger file similar
to the [Ledger](http://ledger-cli.org) command line tool written in C++.

I have been using this tool to track my finances for over a year. I have data
going back over five years.

## Simple Ledger file support

The ledger file this will parse is much simpler than the C++ tool.

Transaction Format:

    <YYYY/MM/dd> <Payee description>
        <Account Name 1>    <Amount 1>
        .
        .
        .
        <Account Name N>    <Amount N>
 
The transaction must balance (the positive amounts must equal negative amounts).
One of the account lines is allowed to have no amount. The amount necessary
to balance the transaction will be added to that account for the transaction.
Amounts must be decimal numbers with a negative(-) sign in front if necessary.

Example transaction:

    2013/01/02 McDonald's #24233 HOUSTON TX
        Expenses:Dining Out:Fast Food        5.60
        Wallet:Cash

A ledger file is assumed to be a list of transactions separated by a new line.


## pkg/ledger

This will parse a ledger file into an array of Transaction structs.
There is also a function get balances for all accounts in the ledger file.

[GoDoc](http://godoc.org/github.com/howeyc/ledger/pkg/ledger)

## cmd/ledger

A very simplistic version of Ledger.
Supports "balance", "register", "print" and "stats" commands.

Example usage:
```sh
    ledger -f ledger.dat bal
    ledger -f ledger.dat bal Cash
    ledger -f ledger.dat reg
    ledger -f ledger.dat reg Food
    ledger -f ledger.dat print
    ledger -f ledger.dat stats
```

## cmd/limport

Using an existing ledger as input to a bayesian classifier, it will attempt to
classify an imported csv of transactions based on payee names and print them in
a ledger file format. 

Attempts to get payee, date, and amount based on headers in the csv file.

Example usage:
```sh
    ledger -f ledger.dat discover discover-recent-transactions.csv
```

In the above example "discover" is the account search string to use to find
the account that all transactions in the csv file should be applied too. The
second account to use for each transaction will be picked based on the
bayesian classification of the payee.

## cmd/lreport

Work in progress. Very bare-bones, I recommend lweb.

Create a chart based on ledger file. Currently plots the following Accounts:
* Assets
* Liabilites
* Net Worth (Assets - Liabilities)
* Income

Output is a 11 x 8.5 inch digital file (png or pdf).

The reported time range can be specified, as well as the frequency of times
plotted on the chart. Optional account filter is also allowed.

Example usage:
```sh
    lreport -f ledger.dat --quarterly --output quarterly-chart.png
    lreport -f ledger.dat -s "2011/01/01" --output stocks-chart.png Stocks
```

## cmd/lweb

Another work in progress.

A website view of the ledger file.

Example usage:
```sh
    lweb -f ledger.dat --port 8080
```
