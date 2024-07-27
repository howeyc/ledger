# Performance

Comparison between various ledger-like applications:

- ledger-go
- [ledger-cli](https://ledger-cli.org)
- [hledger](https://hledger.org)

## Stats

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go stats` | 19.5ms ± 600µs | 18.5ms | 21.6ms | 1.00 |
| `ledger-cli stats` | 141.7ms ± 2.8ms | 136.6ms | 150.4ms | 7.25 ± 0.29 |
| `hledger stats` | 1.5932s ± 74ms | 1.5262s | 1.7565s | 81.48 ± 4.69 |

## Balance

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go bal` | 18ms ± 500µs | 17ms | 19.8ms | 1.00 |
| `ledger-cli bal` | 151.2ms ± 2.4ms | 148.9ms | 159.9ms | 8.36 ± 0.30 |
| `hledger bal` | 1.5484s ± 15.4ms | 1.5258s | 1.5748s | 85.63 ± 2.87 |

## Register

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go reg` | 29.7ms ± 900µs | 27.4ms | 32.4ms | 1.00 |
| `ledger-cli reg` | 1.909s ± 15.6ms | 1.8861s | 1.9304s | 64.08 ± 2.19 |
| `hledger reg` | 2.2709s ± 22.9ms | 2.2352s | 2.3067s | 76.23 ± 2.65 |

## Print

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go print` | 26.6ms ± 1.3ms | 23.7ms | 31.3ms | 1.00 |
| `ledger-cli print` | 280.1ms ± 4ms | 273.2ms | 285.3ms | 10.50 ± 0.55 |
| `hledger print` | 1.8524s ± 18.4ms | 1.8038s | 1.8755s | 69.40 ± 3.60 |

