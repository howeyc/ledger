# Performance

Comparison between various ledger-like applications:

- ledger-go
- [ledger-cli](https://ledger-cli.org)
- [hledger](https://hledger.org)

## Stats

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go stats` | 13.9ms ± 800µs | 11.9ms | 16.4ms | 1.00 |
| `ledger-cli stats` | 163ms ± 1.5ms | 161.1ms | 168.4ms | 11.65 ± 0.70 |
| `hledger stats` | 1.3441s ± 36.5ms | 1.3152s | 1.4253s | 96.06 ± 6.26 |

## Balance

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go bal` | 23.3ms ± 700µs | 21.2ms | 25.6ms | 1.00 |
| `ledger-cli bal` | 151.8ms ± 7.1ms | 139.8ms | 169.9ms | 6.50 ± 0.37 |
| `hledger bal` | 1.3373s ± 10.8ms | 1.315s | 1.3554s | 57.23 ± 1.98 |

## Register

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go reg` | 51.5ms ± 1.1ms | 49.2ms | 55.3ms | 1.00 |
| `ledger-cli reg` | 1.7532s ± 18.1ms | 1.723s | 1.7786s | 34.02 ± 0.85 |
| `hledger reg` | 1.9308s ± 14.3ms | 1.9056s | 1.9462s | 37.46 ± 0.90 |

## Print

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go print` | 40.3ms ± 800µs | 38.4ms | 42.3ms | 1.00 |
| `ledger-cli print` | 238.6ms ± 3.2ms | 234.1ms | 246ms | 5.92 ± 0.15 |
| `hledger print` | 1.5484s ± 10.1ms | 1.525s | 1.5558s | 38.42 ± 0.87 |

