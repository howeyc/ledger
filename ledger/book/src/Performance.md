# Performance

Comparison between various ledger-like applications:

- ledger-go
- [ledger-cli](https://ledger-cli.org)
- [hledger](https://hledger.org)
- [rledger](https://rustledger.github.io)

## Stats

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go stats` | 10.2ms ± 600µs | 8.9ms | 12.5ms | 1.00 |
| `rledger report stats` | 39.8ms ± 1.2ms | 37.8ms | 43.4ms | 3.87 ± 0.26 |
| `ledger-cli stats` | 125ms ± 3.3ms | 118.9ms | 130ms | 12.15 ± 0.79 |
| `hledger stats` | 768.9ms ± 3.3ms | 763.3ms | 774.4ms | 74.69 ± 4.44 |

## Balance

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go bal` | 10.6ms ± 500µs | 9.4ms | 12.7ms | 1.00 |
| `rledger report balances` | 44.9ms ± 1ms | 42.6ms | 48.9ms | 4.22 ± 0.25 |
| `ledger-cli bal` | 124.1ms ± 1.3ms | 120.8ms | 127.1ms | 11.68 ± 0.64 |
| `hledger bal` | 728.3ms ± 3ms | 724.4ms | 735.7ms | 68.48 ± 3.71 |

## Register

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go reg` | 15.2ms ± 900µs | 13.1ms | 18ms | 1.00 |
| `rledger report register` | 131.4ms ± 1.5ms | 129.1ms | 136ms | 8.62 ± 0.53 |
| `hledger reg` | 1.1122s ± 5.3ms | 1.1041s | 1.1232s | 72.91 ± 4.45 |
| `ledger-cli reg` | 1.318s ± 38.2ms | 1.2672s | 1.3738s | 86.41 ± 5.82 |

## Print

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go print` | 14.6ms ± 900µs | 12.4ms | 18ms | 1.00 |
| `rledger format` | 47.5ms ± 1.1ms | 44.6ms | 50ms | 3.25 ± 0.24 |
| `ledger-cli print` | 222.8ms ± 3.8ms | 215.4ms | 228.4ms | 15.22 ± 1.07 |
| `hledger print` | 946.9ms ± 5.3ms | 936.8ms | 954.1ms | 64.68 ± 4.42 |

