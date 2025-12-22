# Keyp vs Redis Performance Comparison

**Generated:** 2025-12-22 15:24:20

## Configuration

| Parameter | Redis | Keyp |
|-----------|-------|------|
| Operations | 1000 | 1000 |
| Clients | 2 | 2 |
| Key Size | 16 bytes | 16 bytes |
| Value Size | 64 bytes | 64 bytes |

## Performance Summary

- **Overall Keyp Performance:** Equivalent
- **Average Performance Ratio:** 0.95x
- **Best Keyp Command:** GET
- **Worst Keyp Command:** TTL

## Detailed Results

| Command | Redis ops/sec | Keyp ops/sec | Ratio | Redis Avg | Keyp Avg | Redis P95 | Keyp P95 |
|---------|---------------|--------------|-------|-----------|----------|-----------|----------|
| SET | 11669 | 11780 | 1.01x | 170.0μs | 168.6μs | 242.4μs | 219.7μs |
| GET | 10237 | 10554 | 1.03x | 194.6μs | 188.7μs | 307.0μs | 259.5μs |
| DEL | 5975 | 5921 | 0.99x | 333.6μs | 336.9μs | 435.0μs | 417.4μs |
| EXPIRE | 6396 | 6183 | 0.97x | 311.9μs | 322.6μs | 429.5μs | 412.6μs |
| TTL | 5633 | 3836 | 0.68x | 354.1μs | 520.6μs | 477.2μs | 646.6μs |
| PERSIST | 4115 | 4109 | 1.00x | 485.3μs | 486.0μs | 619.6μs | 640.3μs |

## Success Rates

| Command | Redis Success Rate | Keyp Success Rate |
|---------|-------------------|------------------|
| SET | 100.0% | 100.0% |
| GET | 100.0% | 100.0% |
| DEL | 100.0% | 100.0% |
| EXPIRE | 100.0% | 100.0% |
| TTL | 100.0% | 100.0% |
| PERSIST | 100.0% | 100.0% |
