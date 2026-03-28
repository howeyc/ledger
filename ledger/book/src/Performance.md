# Performance

Comparison between various ledger-like applications:

- ledger-go
- [ledger-cli](https://ledger-cli.org)
- [hledger](https://hledger.org)

## Stats

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go stats` | 10.4ms ± 600µs | 8.9ms | 12.8ms | 1.00 |
| `ledger-cli stats` | 127ms ± 3.7ms | 119.5ms | 131.1ms | 12.11 ± 0.85 |
| `hledger stats` | 771ms ± 6.6ms | 759.3ms | 784.1ms | 73.46 ± 4.69 |

## Balance

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go bal` | 10.7ms ± 600µs | 9.4ms | 12.7ms | 1.00 |
| `ledger-cli bal` | 124.9ms ± 2.7ms | 119ms | 131.2ms | 11.60 ± 0.75 |
| `hledger bal` | 729ms ± 3.4ms | 724.3ms | 734.4ms | 67.65 ± 4.13 |

## Register

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go reg` | 15.2ms ± 1ms | 12.3ms | 18.1ms | 1.00 |
| `hledger reg` | 1.1036s ± 5.2ms | 1.0977s | 1.1162s | 72.52 ± 4.82 |
| `ledger-cli reg` | 1.3065s ± 43.4ms | 1.2367s | 1.3515s | 85.85 ± 6.37 |

## Print

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go print` | 14.2ms ± 900µs | 11.7ms | 16.9ms | 1.00 |
| `ledger-cli print` | 223.2ms ± 3.8ms | 217.4ms | 229.7ms | 15.69 ± 1.07 |
| `hledger print` | 946.6ms ± 3.7ms | 940.7ms | 954.5ms | 66.51 ± 4.41 |

