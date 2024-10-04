# Performance

Comparison between various ledger-like applications:

- ledger-go
- [ledger-cli](https://ledger-cli.org)
- [hledger](https://hledger.org)

## Stats

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go stats` | 16.9ms ± 700µs | 15.4ms | 19.4ms | 1.00 |
| `ledger-cli stats` | 139.3ms ± 1.8ms | 136ms | 145.5ms | 8.23 ± 0.40 |
| `hledger stats` | 1.5835s ± 22.7ms | 1.5659s | 1.6467s | 93.49 ± 4.54 |

## Balance

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go bal` | 16.2ms ± 800µs | 15.1ms | 18.8ms | 1.00 |
| `ledger-cli bal` | 149.5ms ± 2.1ms | 147.5ms | 157.4ms | 9.19 ± 0.48 |
| `hledger bal` | 1.5783s ± 7.7ms | 1.5656s | 1.5877s | 97.01 ± 4.86 |

## Register

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go reg` | 29ms ± 900µs | 27.2ms | 31.4ms | 1.00 |
| `ledger-cli reg` | 1.9186s ± 17.7ms | 1.8879s | 1.9468s | 65.96 ± 2.20 |
| `hledger reg` | 2.2997s ± 14.6ms | 2.2761s | 2.3275s | 79.06 ± 2.58 |

## Print

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go print` | 25.1ms ± 1.5ms | 22ms | 29.2ms | 1.00 |
| `ledger-cli print` | 281.7ms ± 5.7ms | 275.6ms | 296.1ms | 11.20 ± 0.71 |
| `hledger print` | 1.8827s ± 15.1ms | 1.8546s | 1.905s | 74.83 ± 4.52 |

