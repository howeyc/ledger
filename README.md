[![license](https://img.shields.io/badge/license-ISC-brightgreen.svg)](https://en.wikipedia.org/wiki/ISC_license)
[![GitHub releases](https://img.shields.io/github/tag/howeyc/ledger.svg)](https://github.com/howeyc/ledger/releases)
[![GitHub downloads](https://img.shields.io/github/downloads/howeyc/ledger/total.svg?logo=github&logoColor=lime)](https://github.com/howeyc/ledger/releases)
[![Chat on Libera](https://img.shields.io/badge/chat-libera-blue.svg)](https://matrix.to/#/#plaintextaccounting:libera.chat)
[![Go Report Card](https://goreportcard.com/badge/github.com/howeyc/ledger)](https://goreportcard.com/report/github.com/howeyc/ledger)
[![Go Reference](https://pkg.go.dev/badge/github.com/howeyc/ledger.svg)](https://pkg.go.dev/github.com/howeyc/ledger)
[![Coverage Status](https://coveralls.io/repos/github/howeyc/ledger/badge.svg?branch=master)](https://coveralls.io/github/howeyc/ledger?branch=master)

<div align="center">
 <img src="logo.png" width="50%" height="50%" alt="ledger-logo">
</div>

## Purpose

Ledger is a command line application for plain text accounting. Providing
commands to view balances, register of transactions, importing of CSV files,
exporting of CSV files, and a web interface to view reports, and track
investments. The transaction file format is heavily inspired by
[ledger-cli](https://ledger-cli.org), with a focus on simplicity and
performance.

## Documentation

Head over to https://howeyc.github.io/ledger/

## Performance

Comparison between various ledger-like applications:

- ledger-go
- [ledger-cli](https://ledger-cli.org)
- [hledger](https://hledger.org)

### Stats

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go stats` | 14.2ms ± 500µs | 13.1ms | 17.2ms | 1.00 |
| `ledger-cli stats` | 165.5ms ± 1.3ms | 163.3ms | 169.6ms | 11.58 ± 0.49 |
| `hledger-cli stats` | 1.3275s ± 10.9ms | 1.3051s | 1.3458s | 92.90 ± 3.92 |

### Balance

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go bal` | 23.9ms ± 700µs | 22.9ms | 28.1ms | 1.00 |
| `ledger-cli bal` | 139.5ms ± 1.3ms | 137.5ms | 144.3ms | 5.83 ± 0.18 |
| `hledger-cli bal` | 1.333s ± 8.8ms | 1.3252s | 1.3559s | 55.66 ± 1.71 |

### Register

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go reg` | 52ms ± 1.1ms | 50.5ms | 57.7ms | 1.00 |
| `ledger-cli reg` | 1.7474s ± 20.8ms | 1.7138s | 1.7829s | 33.60 ± 0.84 |
| `hledger-cli reg` | 1.9381s ± 8.8ms | 1.926s | 1.9564s | 37.26 ± 0.84 |

### Print

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go print` | 40.9ms ± 800µs | 39.2ms | 44.2ms | 1.00 |
| `ledger-cli print` | 239.1ms ± 2.8ms | 235.2ms | 246.5ms | 5.83 ± 0.14 |
| `hledger-cli print` | 1.5487s ± 10ms | 1.5254s | 1.556s | 37.78 ± 0.83 |

