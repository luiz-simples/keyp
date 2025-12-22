# Padrões de Código - Projeto Keyp

## Nomenclatura

### Receivers - OBRIGATÓRIO: Nomes Descritivos
```go
// ✅ CORRETO - Nome descritivo
func (server *Server) Start() error
func (storage *LMDBStorage) Get(key []byte) ([]byte, error)
func (config *Config) Validate() error

// ❌ PROIBIDO - Abreviações de uma letra
func (s *Server) Start() error
func (c *Config) Validate() error
func (l *LMDBStorage) Get() error
```

### Variáveis e Funções
```go
// ✅ Bom - Nomes curtos mas claros
func parseCmd(data []byte) (*Command, error)
func execSet(key, val []byte) error
type ConnPool struct{}

// ❌ Evitar
func p(d []byte) (*Command, error)  // muito curto
func executeSetCommand(key, value []byte) error  // muito verboso
func c() error  // sem contexto
```

### Packages e Arquivos
- **Packages**: Nomes curtos, sem underscores ou camelCase
- **Files**: snake_case, um conceito por arquivo
- **Interfaces**: Terminação -er quando apropriado (Reader, Writer)
- **Structs**: PascalCase para exportados, camelCase para privados

## Estruturas de Controle

### OBRIGATÓRIO: Return Early, NUNCA If/Else
```go
// ✅ CORRETO - Return early
func (storage *LMDBStorage) Del(keys ...[]byte) (int, error) {
    err := storage.performDelete(keys)
    if err != nil {
        return 0, err
    }
    
    return len(keys), nil
}

// ❌ PROIBIDO - If/else
func (storage *LMDBStorage) Del(keys ...[]byte) (int, error) {
    err := storage.performDelete(keys)
    if err != nil {
        return 0, err
    } else {  // ← PROIBIDO
        return len(keys), nil
    }
}
```

### OBRIGATÓRIO: If Continue em Loops, NUNCA Else
```go
// ✅ CORRETO - If continue em loop
func (storage *LMDBStorage) processKeys(keys [][]byte) error {
    for _, key := range keys {
        isEmpty := len(key) == 0
        if isEmpty {
            continue  // ← CORRETO
        }
        
        err := storage.processKey(key)
        if err != nil {
            return err
        }
    }
    return nil
}

// ❌ PROIBIDO - If/else em loop
func (storage *LMDBStorage) processKeys(keys [][]byte) error {
    for _, key := range keys {
        if len(key) == 0 {
            continue
        } else {  // ← PROIBIDO
            // código aqui
        }
    }
    return nil
}
```

### OBRIGATÓRIO: Maps ao invés de Switch
```go
// ✅ CORRETO - Map dispatch
type Server struct {
    handlers map[string]func(redcon.Conn, redcon.Command)
}

func (server *Server) setupHandlers() {
    server.handlers = map[string]func(redcon.Conn, redcon.Command){
        "PING": handlePing,
        "ECHO": handleEcho,
        "SET":  handleSet,
    }
}

// ❌ PROIBIDO - Switch statement
func (server *Server) handleCommand(conn redcon.Conn, cmd redcon.Command) {
    switch string(cmd.Args[0]) {  // ← PROIBIDO
    case "PING":
        handlePing(conn, cmd)
    case "ECHO":
        handleEcho(conn, cmd)
    }
}
```

## Métodos vs Funções

### OBRIGATÓRIO: Métodos Apenas Quando Dependem do Estado
```go
// ✅ CORRETO - Depende do estado do struct
func (server *Server) Start() error {
    server.running = true  // Modifica estado
    return server.listener.Listen()  // Usa estado
}

// ✅ CORRETO - Função independente
func handlePing(conn redcon.Conn, cmd redcon.Command) {
    conn.WriteString("PONG")  // Não usa estado de nenhum struct
}

// ❌ PROIBIDO - Método que não usa estado
func (server *Server) handlePing(conn redcon.Conn, cmd redcon.Command) {
    conn.WriteString("PONG")  // Não usa server.*
}
```

**Regra:** Se a função NÃO acessa ou modifica campos do struct, DEVE ser função independente.

## Variáveis Booleanas Explícitas

### OBRIGATÓRIO: Condições extraídas para funcões nomenclaturas descrevendo a condição
```go
// ✅ CORRETO - Variável booleana explícita
func isEmpty(key []byte) bool {
    return len(key) == 0
}
func isExceedsLimit(key []byte) bool {
    return len(key) > MaxKeySize
}
func validateKey(key []byte) error {
    if isEmpty(key) {
        return ErrEmptyKey
    }

    if isExceedsLimit(key) {
        return ErrKeyTooLarge
    }
    
    return nil
}

// ❌ PROIBIDO - Condição inline
func validateKey(key []byte) error {
    if len(key) == 0 {  // ← PROIBIDO - condição não explícita
        return ErrEmptyKey
    }
    return nil
}
```

## Tipos e Interfaces

### OBRIGATÓRIO: any ao invés de interface{}
```go
// ✅ CORRETO - Usar any (Go 1.18+)
var pool = sync.Pool{
    New: func() any {
        return &Command{}
    },
}

func process(value any) error {
    return nil
}

// ❌ PROIBIDO - interface{} legado
var pool = sync.Pool{
    New: func() interface{} {  // ← PROIBIDO
        return &Command{}
    },
}
```

## Comentários

### OBRIGATÓRIO: Zero Comentários
```go
// ✅ CORRETO - Código autoexplicativo
func (storage *LMDBStorage) buildMetaKey(key []byte) []byte {
    metaKey := make([]byte, len(key)+1)
    metaKey[0] = 0xFF
    copy(metaKey[1:], key)
    return metaKey
}

// ❌ PROIBIDO - Qualquer comentário
func (storage *LMDBStorage) buildMetaKey(key []byte) []byte {
    // Create metadata key with prefix ← PROIBIDO
    metaKey := make([]byte, len(key)+1)
    metaKey[0] = 0xFF  // Add prefix ← PROIBIDO
    copy(metaKey[1:], key)
    return metaKey
}
```

**Regra:** ZERO comentários. Código deve ser autoexplicativo através de nomes.

## Organização de Código

### Estrutura de Arquivos
- Um conceito principal por arquivo
- Helpers em arquivos auxiliares quando não dependem de structs
- Métodos privados no mesmo arquivo do struct quando dependem do estado
- Sufixo _test.go para testes

### Dependências do Projeto
**Bibliotecas Obrigatórias:**
- tidwall/redcon (servidor Redis)
- bmatsenyuk/lmdb-go (storage)
- onsi/ginkgo (testes)
- onsi/gomega (matchers)
- leanovate/gopter (property tests)

### Configuração
- Environment variables com prefixo KEYP_
- Arquivos de configuração em YAML ou TOML
- Valores padrão sensatos
- Validação na inicialização

## Checklist de Conformidade

Antes de qualquer commit:
- [ ] Receivers usam nomes descritivos (não `s`, `c`, `l`)
- [ ] Métodos apenas quando dependem do estado do struct
- [ ] Funções independentes quando não usam estado
- [ ] Zero `else` - apenas return early
- [ ] Zero `switch` - apenas maps
- [ ] Condições extraídas para funcões nomenclaturas descrevendo a condição
- [ ] `any` ao invés de `interface{}`
- [ ] Zero comentários
- [ ] Nomes de variáveis curtos mas claros

## Violações comuns a evitar

1. **Receiver de uma letra:** `func (s *Server)` → `func (server *Server)`
2. **Método sem estado:** `func (s *Server) handlePing()` → `func handlePing()`
3. **If/else em função:** `if err != nil {...} else {...}` → `if err != nil {...} return ...`
4. **If/else em loop:** `if condition {...} else {...}` → `if condition { continue } ...`
5. **Switch:** `switch cmd {...}` → `handlers[cmd](...)`
6. **Interface{} legado:** `func(value interface{})` → `func(value any)`
7. **Condição inline:** `if len(key) == 0` → `func isEmpty(key []byte) { return len(key) == 0 } if isEmpty(key)`