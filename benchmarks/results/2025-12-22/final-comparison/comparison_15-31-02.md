# Keyp vs Redis Performance Comparison

**Generated:** 2025-12-22 15:31:02

## Configuration

| Parameter | Redis | Keyp |
|-----------|-------|------|
| Operations | 1000 | 1000 |
| Clients | 2 | 2 |
| Key Size | 16 bytes | 16 bytes |
| Value Size | 64 bytes | 64 bytes |

## Performance Summary

- **Overall Keyp Performance:** Inferior
- **Average Performance Ratio:** 0.40x
- **Best Keyp Command:** GET
- **Worst Keyp Command:** DEL

## Detailed Results

| Command | Redis ops/sec | Keyp ops/sec | Ratio | Redis Avg | Keyp Avg | Redis P95 | Keyp P95 |
|---------|---------------|--------------|-------|-----------|----------|-----------|----------|
| SET | 21233 | 11419 | 0.54x | 92.8μs | 174.1μs | 176.0μs | 250.0μs |
| GET | 20418 | 11229 | 0.55x | 97.0μs | 177.3μs | 153.0μs | 249.0μs |
| DEL | 20525 | 5719 | 0.28x | 96.1μs | 348.5μs | 148.8μs | 472.2μs |
| EXPIRE | 20463 | 6197 | 0.30x | 96.5μs | 321.8μs | 148.1μs | 421.2μs |
| TTL | 13791 | 5674 | 0.41x | 144.0μs | 351.6μs | 220.0μs | 446.5μs |
| PERSIST | 13427 | 4033 | 0.30x | 147.8μs | 495.0μs | 220.2μs | 638.5μs |

## Success Rates

| Command | Redis Success Rate | Keyp Success Rate |
|---------|-------------------|------------------|
| SET | 100.0% | 100.0% |
| GET | 100.0% | 100.0% |
| DEL | 100.0% | 100.0% |
| EXPIRE | 100.0% | 100.0% |
| TTL | 100.0% | 100.0% |
| PERSIST | 100.0% | 100.0% |
