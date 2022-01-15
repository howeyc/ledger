# Running ledger

Starting ledger provides us with a list of all the commands that are available.

```sh
ledger
```

This produces the following output.

```
Plain text accounting

Usage:
  ledger [command]

Available Commands:
  balance     Print account balances
  completion  generate the autocompletion script for the specified shell
  equity      Print account equity as transaction
  help        Help about any command
  import      Import transactions from csv to ledger format
  lint        Check ledger for errors
  print       Print transactions in ledger file format
  register    Print register of transactions
  stats       A small report of transaction stats
  version     Version of ledger
  web         Web service

Flags:
  -f, --file string   ledger file (default is $LEDGER_FILE) (default "")
  -h, --help          help for ledger

Use "ledger [command] --help" for more information about a command.
```

In order to run any command we must specify the ledger file. This is done with
either the **-f** or **--file** flag. However, since this needs to be included
so often, it can also be specified via the environment variable
**LEDGER_FILE**.

It is encouraged to setup this **LEDGER_FILE** to require less typing every time
a command is run.
