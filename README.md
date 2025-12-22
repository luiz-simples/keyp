# Keyp

Keyp é um servidor de armazenamento chave-valor compatível com o protocolo Redis, implementado em Go com LMDB como backend de persistência.

## Características

- Protocolo Redis compatível (SET, GET, DEL, PING)
- Persistência com LMDB (Lightning Memory-Mapped Database)
- Alta performance com baixa latência
- Testes unitários, de integração e property-based
- Benchmarks de performance

## Requisitos

- Go 1.25+
- LMDB (instalado automaticamente via go mod)

## Instalação

```bash
make deps
make build
```

## Uso

### Iniciar o servidor

```bash
make run
```

Ou com opções customizadas:

```bash
./bin/keyp -host localhost -port 6380 -data-dir ./data
```

### Opções de linha de comando

- `-host`: Host para bind (padrão: localhost)
- `-port`: Porta para escutar (padrão: 6380)
- `-data-dir`: Diretório para dados LMDB (padrão: ./data)

## Testes

### Executar todos os testes

```bash
make test
```

### Testes unitários

```bash
make test-unit
```

### Testes de integração

```bash
make test-integration
```

### Property-based tests

```bash
make test-property
```

### Benchmarks

```bash
make benchmark
```

## Comandos Suportados

### PING

```bash
redis-cli -p 6380 PING
# PONG

redis-cli -p 6380 PING "hello"
# "hello"
```

### SET

```bash
redis-cli -p 6380 SET mykey "myvalue"
# OK
```

### GET

```bash
redis-cli -p 6380 GET mykey
# "myvalue"

redis-cli -p 6380 GET nonexistent
# (nil)
```

### DEL

```bash
redis-cli -p 6380 DEL mykey
# (integer) 1

redis-cli -p 6380 DEL key1 key2 key3
# (integer) 3
```

## Arquitetura

```
keyp/
├── cmd/keyp/           # Aplicação principal
├── internal/
│   ├── server/         # Servidor Redis protocol
│   └── storage/        # LMDB storage layer
├── bin/                # Binários compilados
└── data/               # Dados LMDB
```

## Performance

Benchmarks no Apple M1 Pro:

```
BenchmarkLMDBStorage/Set-10            16369    76968 ns/op    127 B/op    5 allocs/op
BenchmarkLMDBStorage/Get-10           973527     1337 ns/op    144 B/op    5 allocs/op
BenchmarkLMDBStorage/Del-10            15110    76160 ns/op     80 B/op    3 allocs/op
BenchmarkLMDBStorage/SetGetDel-10       7514   155147 ns/op    351 B/op   13 allocs/op
```

## Desenvolvimento

### Estrutura de código

O projeto segue padrões rigorosos de código Go:

- Receivers com nomes descritivos (não abreviações)
- Return early (sem else)
- Maps ao invés de switch
- Funções independentes quando não usam estado
- Zero comentários (código autoexplicativo)

### Comandos úteis

```bash
make format      # Formatar código
make lint        # Executar linter
make clean       # Limpar arquivos gerados
make all         # Build completo com testes
```

## Docker

```bash
make docker-build
make docker-run
```

## Licença

MIT
