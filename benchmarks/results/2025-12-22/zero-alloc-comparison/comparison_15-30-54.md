# Keyp vs Redis Performance Comparison

**Generated:** 2025-12-22 15:30:54

## Configuration

| Parameter | Redis | Keyp |
|-----------|-------|------|
| Operations | 1000 | 1000 |
| Clients | 2 | 2 |
| Key Size | 16 bytes | 16 bytes |
| Value Size | 64 bytes | 64 bytes |

## Performance Summary

- **Overall Keyp Performance:** Equivalent
- **Average Performance Ratio:** 1.08x
- **Best Keyp Command:** TTL
- **Worst Keyp Command:** DEL

## Detailed Results

| Command | Redis ops/sec | Keyp ops/sec | Ratio | Redis Avg | Keyp Avg | Redis P95 | Keyp P95 |
|---------|---------------|--------------|-------|-----------|----------|-----------|----------|
| SET | 11780 | 11419 | 0.97x | 168.6μs | 174.1μs | 219.7μs | 250.0μs |
| GET | 10554 | 11229 | 1.06x | 188.7μs | 177.3μs | 259.5μs | 249.0μs |
| DEL | 5921 | 5719 | 0.97x | 336.9μs | 348.5μs | 417.4μs | 472.2μs |
| EXPIRE | 6183 | 6197 | 1.00x | 322.6μs | 321.8μs | 412.6μs | 421.2μs |
| TTL | 3836 | 5674 | 1.48x | 520.6μs | 351.6μs | 646.6μs | 446.5μs |
| PERSIST | 4109 | 4033 | 0.98x | 486.0μs | 495.0μs | 640.3μs | 638.5μs |

## Success Rates

| Command | Redis Success Rate | Keyp Success Rate |
|---------|-------------------|------------------|
| SET | 100.0% | 100.0% |
| GET | 100.0% | 100.0% |
| DEL | 100.0% | 100.0% |
| EXPIRE | 100.0% | 100.0% |
| TTL | 100.0% | 100.0% |
| PERSIST | 100.0% | 100.0% |
