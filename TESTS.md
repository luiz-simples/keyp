# Keyp - Estratégia de Testes

> **Documentação Relacionada**: [README.md](README.md) | [ARCHITECTURE.md](ARCHITECTURE.md)

## Visão Geral

Este documento descreve a estratégia completa de testes do projeto Keyp, implementada seguindo os padrões de código definidos no projeto.

## Cobertura Atual

### ✅ Package `internal/service` (Completo)

Implementação completa de 4 tipos de testes com isolamento total e paralelização segura.

#### 🧪 Unit Tests
- **Arquivo**: `internal/service/unit_test.go`
- **Framework**: Ginkgo + Gomega + GoMock
- **Cobertura**: Comandos PING, SET, GET, DEL com mocks
- **Cenários**: Sucesso, erro, contexto cancelado, validação
- **Mocks**: Gerados com `mockgen` para `domain.Persister`

#### 🔗 Integration Tests  
- **Arquivo**: `internal/service/integration_test.go`
- **Framework**: Ginkgo + Gomega + go-redis
- **Cobertura**: Servidor Redis real com cliente go-redis
- **Cenários**: Operações básicas, concorrência, valores grandes
- **Protocolo**: Compatibilidade completa com Redis

#### 🎯 Property-Based Tests
- **Arquivo**: `internal/service/property_test.go` 
- **Framework**: Ginkgo + Gomega + Gopter
- **Cobertura**: Propriedades fundamentais (SET-GET, DEL, etc.)
- **Cenários**: 100 testes por propriedade com dados aleatórios
- **Validação**: Invariantes e comportamentos esperados

#### ⚡ Performance Tests
- **Arquivo**: `internal/service/performance_test.go`
- **Framework**: Ginkgo + Benchmarks Go nativos
- **Cobertura**: Métricas de performance e validação de tempo
- **Cenários**: SET, GET, DEL, PING, operações mistas
- **Benchmarks**: Métricas precisas de ns/op

### ✅ Package `internal/storage` (Completo)

Implementação completa de 4 tipos de testes para o sistema de persistência LMDB com isolamento total.

#### 🧪 Unit Tests (26 testes)
- **Arquivo**: `internal/storage/unit_test.go`
- **Framework**: Ginkgo + Gomega
- **Cobertura**: Todas as operações LMDB (Set, Get, Del, TTL, Expire, Persist)
- **Cenários**: Criação de cliente, isolamento de databases, tratamento de erros
- **Validação**: Chaves vazias, valores grandes, contextos cancelados

#### 🔗 Integration Tests (12 testes)
- **Arquivo**: `internal/storage/integration_test.go`
- **Framework**: Ginkgo + Gomega
- **Cobertura**: Instâncias reais do LMDB com operações concorrentes
- **Cenários**: Múltiplas goroutines, isolamento entre databases, dados grandes
- **Validação**: Thread-safety, TTL, timeout e cancelamento de contexto

#### 🎯 Property-Based Tests (10 testes)
- **Arquivo**: `internal/storage/property_test.go`
- **Framework**: Ginkgo + Gomega + Gopter
- **Cobertura**: Invariantes do storage (Set-Get, Set-Delete, TTL, Persist)
- **Cenários**: 1000 testes por propriedade (100 execuções × 10 propriedades)
- **Validação**: Isolamento entre databases, idempotência, consistência

#### ⚡ Performance Tests (12 testes + benchmarks)
- **Arquivo**: `internal/storage/performance_test.go`
- **Framework**: Ginkgo + gmeasure + Benchmarks Go
- **Cobertura**: Performance individual e em lote, concorrência, throughput
- **Cenários**: Operações individuais, batch operations, dados grandes
- **Benchmarks**: Métricas precisas de ns/op para LMDB

### 🔄 Próximos Packages

- `internal/app` - Planejado  
- `cmd/keyp` - Planejado

## Arquitetura de Testes

### Isolamento e Paralelização

#### ✅ Características Implementadas

- **Diretórios Únicos**: Cada teste usa diretório temporário único
- **Limpeza Automática**: Diretórios são removidos após cada teste
- **Paralelização Segura**: Testes podem rodar em paralelo sem conflitos
- **Isolamento Total**: Nenhum teste interfere com outro

#### 📁 Padrão de Diretórios

```
/tmp/keyp-{tipo}-{pid}-{timestamp}/
```

Exemplos:
- `/tmp/keyp-integration-12345-1766756124000/`
- `/tmp/keyp-property-12345-1766756125000/`
- `/tmp/keyp-bench-set-12345-1766756126000/`

#### 🔄 Limpeza Automática

- **BeforeEach**: Cria diretório único
- **AfterEach**: Remove diretório e fecha storage
- **Benchmarks**: Usa `defer` para limpeza garantida

## Execução dos Testes

### Comandos Principais

#### Todos os Testes (Paralelo - Recomendado)
```bash
# Service package
ginkgo -p ./internal/service

# Storage package  
ginkgo -p ./internal/storage

# Todos os packages
ginkgo -p ./internal/...
```

#### Todos os Testes (Sequencial)
```bash
# Service package
go test ./internal/service -v

# Storage package
go test ./internal/storage -v

# Todos os packages
go test ./internal/... -v
```

#### Por Tipo de Teste
```bash
# Unit Tests
go test ./internal/service -v --ginkgo.label-filter="unit"
go test ./internal/storage -v --ginkgo.label-filter="unit"

# Integration Tests  
go test ./internal/service -v --ginkgo.label-filter="integration"
go test ./internal/storage -v --ginkgo.label-filter="integration"

# Property Tests
go test ./internal/service -v --ginkgo.label-filter="property"
go test ./internal/storage -v --ginkgo.label-filter="property"

# Performance Tests
go test ./internal/service -v --ginkgo.label-filter="performance"
go test ./internal/storage -v --ginkgo.label-filter="performance"
```

#### Benchmarks
```bash
# Service benchmarks
go test ./internal/service -bench=. -run=^$

# Storage benchmarks
go test ./internal/storage -bench=. -run=^$

# Benchmark específico
go test ./internal/service -bench=BenchmarkHandlerSET -run=^$
go test ./internal/storage -bench=BenchmarkStorageSet -run=^$
```

### Resultados de Performance

#### Service Package (Apple M1 Pro)

```
BenchmarkHandlerSET-10      865309    1304 ns/op
BenchmarkHandlerGET-10     1000000    1181 ns/op  
BenchmarkHandlerDEL-10      409984    3026 ns/op
BenchmarkHandlerPING-10   19510212      62 ns/op
BenchmarkHandlerMixed-10    277227    4378 ns/op
```

#### Storage Package (Apple M1 Pro)

```
BenchmarkStorageSet-10     1025377    1142 ns/op
BenchmarkStorageGet-10     1000000    1096 ns/op
BenchmarkStorageDel-10      442477    2766 ns/op
BenchmarkStorageMixed-10    305694    3919 ns/op
```

#### Validações de Performance

**Service Layer:**
- **SET**: < 1 segundo para 1000 operações
- **GET**: < 1 segundo para 1000 operações  
- **PING**: < 0.5 segundos para 10000 operações
- **Mixed**: < 3 segundos para 1000 operações completas

**Storage Layer:**
- **SET**: < 1 segundo para 1000 operações LMDB
- **GET**: < 1 segundo para 1000 operações LMDB
- **DEL**: < 3 segundos para 1000 operações LMDB
- **Mixed**: < 4 segundos para 1000 operações completas

#### Paralelização

**Service Package:**
```
Sequencial: 2.8s (37 specs)
Paralelo:   1.7s (37 specs) - 40% mais rápido
Processos:  9 paralelos
```

**Storage Package:**
```
Sequencial: 7.6s (60 specs)
Paralelo:   4.2s (60 specs) - 45% mais rápido  
Processos:  10 paralelos
```

## Padrões de Código nos Testes

### Conformidade com Steering Rules

Os testes seguem rigorosamente os padrões definidos em `.kiro/steering/code-standards.md`:

- ✅ **Zero comentários** - Nomes descritivos
- ✅ **Return early** - Sem `if/else`
- ✅ **Funções extraídas** - `hasError()`, `isEmpty()`
- ✅ **Receivers descritivos** - `handler`, não `h`
- ✅ **Commits semânticos** - `test: add comprehensive coverage`

### Estrutura dos Arquivos

**Service Package:**
```
internal/service/
├── service_test.go      # Suite principal + utilitários
├── mocks_test.go        # Mocks gerados (mockgen)
├── unit_test.go         # Testes unitários
├── integration_test.go  # Testes de integração
├── property_test.go     # Testes baseados em propriedades
└── performance_test.go  # Testes de performance
```

**Storage Package:**
```
internal/storage/
├── storage_test.go      # Suite principal + utilitários
├── unit_test.go         # Testes unitários (26 specs)
├── integration_test.go  # Testes de integração (12 specs)
├── property_test.go     # Testes baseados em propriedades (10 specs)
└── performance_test.go  # Testes de performance (12 specs + benchmarks)
```

### Utilitários Centralizados

```go
func createUniqueTestDir(prefix string) string
func cleanupTestDir(dir string)
```

## Dependências de Teste

### Frameworks Principais

- `github.com/onsi/ginkgo/v2` - Framework de testes BDD
- `github.com/onsi/gomega` - Matchers para assertions
- `go.uber.org/mock` - Geração de mocks
- `github.com/leanovate/gopter` - Property-based testing

### Dependências de Integração

- `github.com/redis/go-redis/v9` - Cliente Redis
- `github.com/tidwall/redcon` - Servidor Redis compatível

### Geração de Mocks

```bash
mockgen -source=internal/domain/types.go -destination=internal/service/mocks_test.go -package=service_test
```

## Integração com CI/CD

### Comandos Recomendados

```bash
# Verificação rápida - todos os packages
go test ./internal/... -v

# Verificação completa com paralelização
ginkgo -p ./internal/...

# Benchmarks para métricas
go test ./internal/service -bench=. -run=^$ -benchmem
go test ./internal/storage -bench=. -run=^$ -benchmem

# Por package específico
go test ./internal/service -v --ginkgo.v
go test ./internal/storage -v --ginkgo.v
```

### Métricas de Qualidade

**Service Package:**
- **Cobertura**: 100% dos comandos principais
- **Specs**: 37 testes em 4 tipos
- **Isolamento**: Total entre execuções
- **Performance**: Benchmarks automatizados

**Storage Package:**
- **Cobertura**: 100% das operações LMDB
- **Specs**: 60 testes em 4 tipos
- **Isolamento**: Total entre execuções
- **Performance**: Benchmarks automatizados

**Total do Projeto:**
- **Specs**: 97 testes (37 service + 60 storage)
- **Property Tests**: 1600+ execuções (100 × 16 propriedades)
- **Benchmarks**: 9 benchmarks diferentes
- **Cobertura**: 2 packages completos

## Roadmap de Testes

### Próximas Implementações

1. **`internal/app`** - Testes de servidor e configuração
2. **`cmd/keyp`** - Testes de CLI e inicialização
3. **End-to-End** - Testes completos do sistema

### Melhorias Futuras

- **Cobertura de código** automatizada
- **Testes de carga** com múltiplos clientes
- **Testes de falha** e recuperação
- **Profiling** automatizado
- **Testes de TTL** com expiração real (storage)

---

> **Próximos Passos**: Consulte [ARCHITECTURE.md](ARCHITECTURE.md) para entender a estrutura do sistema e [README.md](README.md) para instruções de uso.