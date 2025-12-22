# Keyp vs Redis Benchmarks

Este diretório contém benchmarks comparativos entre Keyp e Redis oficial.

## Estrutura

```
benchmarks/
├── cmd/
│   └── benchmark/          # Programa principal de benchmark
├── results/
│   └── YYYY-MM-DD/        # Resultados organizados por data
│       ├── keyp/          # Resultados do Keyp
│       ├── redis/         # Resultados do Redis
│       └── comparison/    # Tabelas comparativas
├── scripts/               # Scripts de automação
└── docker/               # Configurações Docker
```

## Comandos Testados

- SET: Inserção de chaves
- GET: Recuperação de chaves  
- DEL: Remoção de chaves
- EXPIRE: Definição de TTL
- TTL: Consulta de TTL
- PERSIST: Remoção de TTL

## Como Executar

```bash
# Executar benchmark completo
make benchmark

# Executar apenas Keyp
make benchmark-keyp

# Executar apenas Redis
make benchmark-redis

# Gerar relatório comparativo
make benchmark-report
```

## Métricas Coletadas

- Operações por segundo (ops/sec)
- Latência média (ms)
- Latência P95 (ms)
- Latência P99 (ms)
- Uso de memória (MB)
- Uso de CPU (%)