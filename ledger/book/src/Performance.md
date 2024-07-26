# Performance

Comparison between various ledger-like applications:

- ledger-go
- [ledger-cli](https://ledger-cli.org)
- [hledger](https://hledger.org)

## Stats

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go stats` | 19.8ms ± 1.5ms | 18.5ms | 33.1ms | 1.00 |
| `ledger-cli stats` | 138.9ms ± 1.4ms | 136.7ms | 143.7ms | 7.01 ± 0.55 |
| `hledger stats` | 1.5407s ± 15.9ms | 1.5136s | 1.5736s | 77.77 ± 6.08 |

## Balance

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go bal` | 18.2ms ± 500µs | 17.3ms | 19.5ms | 1.00 |
| `ledger-cli bal` | 150.3ms ± 2.3ms | 146.6ms | 158.7ms | 8.22 ± 0.29 |
| `hledger bal` | 1.5499s ± 15.3ms | 1.533s | 1.5741s | 84.78 ± 2.76 |

## Register

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go reg` | 46ms ± 1ms | 43.8ms | 48.7ms | 1.00 |
| `ledger-cli reg` | 1.9039s ± 8.5ms | 1.8852s | 1.9154s | 41.31 ± 0.99 |
| `hledger reg` | 2.2478s ± 18.1ms | 2.2169s | 2.2736s | 48.77 ± 1.21 |

## Print

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go print` | 27.1ms ± 1.2ms | 24.1ms | 29.6ms | 1.00 |
| `ledger-cli print` | 280.5ms ± 5.7ms | 274ms | 292.4ms | 10.33 ± 0.53 |
| `hledger print` | 1.8482s ± 15.6ms | 1.8234s | 1.8737s | 68.03 ± 3.23 |

