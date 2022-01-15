# Reports

The web service included in `ledger` allows for the configuration of many types
of different reports, charts, and calculations.

Lets try an example configuration.

`$ cat reports.toml`

## Expenses

This is a pie chart showing the spending per Expense account.

![expenses pie chart](webshots/report-expenses.png)

## Savings

This report calculates a pseudo account "Savings" based on *Income - Expenses*
over time and shows how much money has been saved per month.

![savings bar chart](webshots/report-savings.png)

## Net Worth

Graph Assets against Liabilities.

![net worth line chart](webshots/report-networth.png)

