# Keyp vs Redis Performance Comparison

**Generated:** 2025-12-22 15:16:17

## Configuration

| Parameter | Redis | Keyp |
|-----------|-------|------|
| Operations | 1000 | 1000 |
| Clients | 2 | 2 |
| Key Size | 16 bytes | 16 bytes |
| Value Size | 64 bytes | 64 bytes |

## Performance Summary

- **Overall Keyp Performance:** Inferior
- **Average Performance Ratio:** 0.39x
- **Best Keyp Command:** SET
- **Worst Keyp Command:** DEL

## Detailed Results

| Command | Redis ops/sec | Keyp ops/sec | Ratio | Redis Avg | Keyp Avg | Redis P95 | Keyp P95 |
|---------|---------------|--------------|-------|-----------|----------|-----------|----------|
| SET | 21233 | 11669 | 0.55x | 92.8μs | 170.0μs | 176.0μs | 242.4μs |
| GET | 20418 | 10237 | 0.50x | 97.0μs | 194.6μs | 153.0μs | 307.0μs |
| DEL | 20525 | 5975 | 0.29x | 96.1μs | 333.6μs | 148.8μs | 435.0μs |
| EXPIRE | 20463 | 6396 | 0.31x | 96.5μs | 311.9μs | 148.1μs | 429.5μs |
| TTL | 13791 | 5633 | 0.41x | 144.0μs | 354.1μs | 220.0μs | 477.2μs |
| PERSIST | 13427 | 4115 | 0.31x | 147.8μs | 485.3μs | 220.2μs | 619.6μs |

## Success Rates

| Command | Redis Success Rate | Keyp Success Rate |
|---------|-------------------|------------------|
| SET | 100.0% | 100.0% |
| GET | 100.0% | 100.0% |
| DEL | 100.0% | 100.0% |
| EXPIRE | 100.0% | 100.0% |
| TTL | 100.0% | 100.0% |
| PERSIST | 100.0% | 100.0% |
