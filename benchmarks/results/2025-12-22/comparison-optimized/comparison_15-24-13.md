# Keyp vs Redis Performance Comparison

**Generated:** 2025-12-22 15:24:13

## Configuration

| Parameter | Redis | Keyp |
|-----------|-------|------|
| Operations | 1000 | 1000 |
| Clients | 2 | 2 |
| Key Size | 16 bytes | 16 bytes |
| Value Size | 64 bytes | 64 bytes |

## Performance Summary

- **Overall Keyp Performance:** Inferior
- **Average Performance Ratio:** 0.37x
- **Best Keyp Command:** SET
- **Worst Keyp Command:** TTL

## Detailed Results

| Command | Redis ops/sec | Keyp ops/sec | Ratio | Redis Avg | Keyp Avg | Redis P95 | Keyp P95 |
|---------|---------------|--------------|-------|-----------|----------|-----------|----------|
| SET | 21233 | 11780 | 0.55x | 92.8μs | 168.6μs | 176.0μs | 219.7μs |
| GET | 20418 | 10554 | 0.52x | 97.0μs | 188.7μs | 153.0μs | 259.5μs |
| DEL | 20525 | 5921 | 0.29x | 96.1μs | 336.9μs | 148.8μs | 417.4μs |
| EXPIRE | 20463 | 6183 | 0.30x | 96.5μs | 322.6μs | 148.1μs | 412.6μs |
| TTL | 13791 | 3836 | 0.28x | 144.0μs | 520.6μs | 220.0μs | 646.6μs |
| PERSIST | 13427 | 4109 | 0.31x | 147.8μs | 486.0μs | 220.2μs | 640.3μs |

## Success Rates

| Command | Redis Success Rate | Keyp Success Rate |
|---------|-------------------|------------------|
| SET | 100.0% | 100.0% |
| GET | 100.0% | 100.0% |
| DEL | 100.0% | 100.0% |
| EXPIRE | 100.0% | 100.0% |
| TTL | 100.0% | 100.0% |
| PERSIST | 100.0% | 100.0% |
