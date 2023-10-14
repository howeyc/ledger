# Performance

Comparison between various ledger-like applications:

- ledger-go
- [ledger-cli](https://ledger-cli.org)
- [hledger](https://hledger.org)

## Stats

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go stats` | 14.1ms ± 800µs | 12ms | 16.9ms | 1.00 |
| `ledger-cli stats` | 167.2ms ± 1.3ms | 164.1ms | 170.3ms | 11.84 ± 0.73 |
| `hledger stats` | 1.3784s ± 10.5ms | 1.3568s | 1.3974s | 97.63 ± 6.01 |

## Balance

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go bal` | 13.4ms ± 700µs | 11.7ms | 16ms | 1.00 |
| `ledger-cli bal` | 143.6ms ± 1.7ms | 141.3ms | 148.9ms | 10.71 ± 0.60 |
| `hledger bal` | 1.3742s ± 9.1ms | 1.357s | 1.3876s | 102.39 ± 5.65 |

## Register

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go reg` | 45.4ms ± 900µs | 43.7ms | 48.3ms | 1.00 |
| `ledger-cli reg` | 1.7282s ± 9.1ms | 1.7129s | 1.748s | 38.04 ± 0.81 |
| `hledger reg` | 1.9924s ± 10.1ms | 1.9784s | 2.0087s | 43.85 ± 0.94 |

## Print

| Command | Mean | Min | Max | Relative |
|:---|---:|---:|---:|---:|
| `ledger-go print` | 32.7ms ± 900µs | 30.9ms | 36.5ms | 1.00 |
| `ledger-cli print` | 244.4ms ± 2ms | 241.7ms | 249ms | 7.47 ± 0.23 |
| `hledger print` | 1.6554s ± 15.1ms | 1.6176s | 1.6689s | 50.62 ± 1.60 |

