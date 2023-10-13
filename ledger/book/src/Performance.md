# Performance

Comparison between various ledger-like applications:

- ledger-go
- [ledger-cli](https://ledger-cli.org)
- [hledger](https://hledger.org)

## Stats

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go stats` | 13.9ms ± 700µs | 12.2ms | 16.1ms | 1.00 |
| `ledger-cli stats` | 166.7ms ± 1.6ms | 164ms | 172ms | 11.93 ± 0.65 |
| `hledger stats` | 1.3901s ± 6.3ms | 1.3771s | 1.3972s | 99.42 ± 5.32 |

## Balance

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go bal` | 20.9ms ± 500µs | 19.7ms | 23.3ms | 1.00 |
| `ledger-cli bal` | 143.6ms ± 1.9ms | 141.3ms | 150.1ms | 6.85 ± 0.22 |
| `hledger bal` | 1.3921s ± 8.1ms | 1.3768s | 1.4074s | 66.36 ± 1.93 |

## Register

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go reg` | 45.4ms ± 1ms | 43.8ms | 49.3ms | 1.00 |
| `ledger-cli reg` | 1.724s ± 11ms | 1.7066s | 1.7445s | 37.90 ± 0.91 |
| `hledger reg` | 1.996s ± 8.7ms | 1.978s | 2.0081s | 43.88 ± 1.03 |

## Print

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go print` | 32.5ms ± 1ms | 30.4ms | 36.2ms | 1.00 |
| `ledger-cli print` | 244.5ms ± 1.9ms | 242.3ms | 249ms | 7.52 ± 0.25 |
| `hledger print` | 1.6632s ± 8.1ms | 1.6481s | 1.6783s | 51.14 ± 1.67 |

