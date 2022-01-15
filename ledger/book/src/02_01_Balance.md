# Balance

Run `ledger -f ledger.dat bal` to see a balance report.

```
Assets                                                                   2920.00
Assets:Bank                                                              2820.00
Assets:Bank:Checking                                                     2620.00
Assets:Bank:Savings                                                       200.00
Assets:Cash                                                               100.00
Assets:Cash:Wallet                                                        100.00
Equity                                                                  -1000.00
Equity:Opening Balances                                                 -1000.00
Expenses                                                                  100.00
Expenses:Books                                                             20.00
Expenses:Food                                                              80.00
Expenses:Food:Groceries                                                    80.00
Income                                                                  -2000.00
Income:Salary                                                           -2000.00
Liabilities                                                               -20.00
Liabilities:MasterCard                                                    -20.00
--------------------------------------------------------------------------------
                                                                            0.00
```

## Net Worth

You can show specific accounts by applying a filter, which is case senstive.
For example, let's get our net worth,
run `ledger -f ledger.dat bal Assets Liabilities`

```
Assets                                                                   2920.00
Assets:Bank                                                              2820.00
Assets:Bank:Checking                                                     2620.00
Assets:Bank:Savings                                                       200.00
Assets:Cash                                                               100.00
Assets:Cash:Wallet                                                        100.00
Liabilities                                                               -20.00
Liabilities:MasterCard                                                    -20.00
--------------------------------------------------------------------------------
                                                                         2900.00
```
