# Padrões de Código - Projeto Keyp

## 🚨 REGRA FUNDAMENTAL
**SEMPRE CONFIRMAR ANTES DE EXECUTAR** - Perguntar ao usuário antes de qualquer ação (modificar código, executar comandos, criar arquivos)

## ⚡ REGRAS CRÍTICAS (Referência Rápida)

### OBRIGATÓRIO
- ✅ Receivers descritivos: `func (server *Server)` NUNCA `func (s *Server)`
- ✅ Funções independentes em `utils.go` - métodos no arquivo principal
- ✅ Return early - ZERO `if/else`
- ✅ Maps dispatch - ZERO `switch`
- ✅ Condições extraídas: `isEmpty(key)` NUNCA `len(key) == 0`
- ✅ Erros extraídos: `hasError(err)` NUNCA `err != nil`
- ✅ `any` NUNCA `interface{}`
- ✅ ZERO comentários

### PROIBIDO
- ❌ `if/else` statements
- ❌ `switch` statements  
- ❌ Condições inline
- ❌ Comparações de erro inline
- ❌ Receivers de uma letra
- ❌ Métodos que não usam estado
- ❌ Comentários

## 🔄 TRANSFORMAÇÕES OBRIGATÓRIAS

### Receivers
```go
❌ func (s *Server) Start() error
✅ func (server *Server) Start() error
```

### Condições
```go
❌ if len(key) == 0 { return ErrEmpty }
✅ if isEmpty(key) { return ErrEmpty }

❌ if err != nil { return err }
✅ if hasError(err) { return err }
```

### Controle de Fluxo
```go
❌ if condition { action() } else { other() }
✅ if condition { action(); return }
   other()

❌ switch cmd { case "SET": handleSet() }
✅ handlers[cmd](conn, cmd)
```

### Métodos vs Funções
```go
❌ func (s *Server) handlePing() // não usa s.*
✅ func handlePing() // função independente

✅ func (server *Server) handleSet() // usa server.storage
```

## 📁 ORGANIZAÇÃO DE ARQUIVOS

### Estrutura Obrigatória
```
package/
├── main.go          # Struct principal + métodos que usam estado
├── utils.go         # Funções independentes (TODAS)
└── *_test.go        # Testes
```

### Regras de Separação
- **`utils.go`**: TODAS as funções que NÃO acessam campos de struct
- **Arquivo principal**: APENAS métodos que acessam/modificam estado
- **Funções obrigatórias em utils.go**:
  ```go
  func hasError(err error) bool { return err != nil }
  func noError(err error) bool { return err == nil }
  func isEmpty(data []byte) bool { return len(data) == 0 }
  ```

## 📝 COMMITS SEMÂNTICOS

### Tipos Obrigatórios
- **feat**: Nova funcionalidade ou comando
- **fix**: Correção de bug ou erro
- **refactor**: Refatoração sem mudança de comportamento
- **test**: Adição ou modificação de testes
- **docs**: Documentação ou README
- **style**: Formatação, lint, organização de código

### Formato Obrigatório
```
tipo: descrição em imperativo minúsculo
```

### Regras de Escrita
- ✅ Imperativo: "add", "fix", "refactor" (não "added", "fixed")
- ✅ Minúsculo: "add set command" (não "Add SET Command")
- ✅ Sem ponto final: "fix memory leak" (não "fix memory leak.")
- ✅ Máximo 50 caracteres na linha de título
- ✅ Descrição clara e específica

### Exemplos Práticos
```bash
feat: add SET command handler
feat: implement DEL operation with multiple keys
fix: resolve memory leak in LMDB storage
fix: handle empty keys in validation
refactor: extract magic numbers to constants
refactor: separate commands into individual files
test: add property tests for storage operations
test: implement integration tests for server
docs: update README with installation guide
style: organize functions into utils.go files
```

### Commits Compostos (Quando Necessário)
```bash
feat: add GET command with error handling
refactor: extract validation functions to utils
test: add unit tests for new command handlers
```

## 🔧 DEPENDÊNCIAS E CONFIGURAÇÃO

### Bibliotecas Obrigatórias
```go
"github.com/PowerDNS/lmdb-go/lmdb"  // Storage LMDB
"github.com/tidwall/redcon"         // Servidor Redis
"github.com/onsi/ginkgo/v2"         // Framework testes
"github.com/onsi/gomega"            // Matchers testes
"github.com/leanovate/gopter"       // Property-based tests
```

### Configuração
- Environment variables: prefixo `KEYP_`
- Arquivos: YAML ou TOML
- Validação na inicialização

## ✅ CHECKLIST DE CONFORMIDADE

Antes de qualquer commit:
- [ ] **CONFIRMAÇÃO**: Perguntei ao usuário antes de executar
- [ ] **RECEIVERS**: Nomes descritivos (não `s`, `c`, `l`)
- [ ] **MÉTODOS**: Apenas quando dependem do estado
- [ ] **FUNÇÕES**: Independentes em `utils.go`
- [ ] **CONTROLE**: Zero `else` - apenas return early
- [ ] **DISPATCH**: Zero `switch` - apenas maps
- [ ] **CONDIÇÕES**: Extraídas para funções nomeadas
- [ ] **ERROS**: Usando `hasError()` e `noError()`
- [ ] **TIPOS**: `any` ao invés de `interface{}`
- [ ] **COMENTÁRIOS**: Zero comentários
- [ ] **DEPENDÊNCIAS**: Bibliotecas corretas
- [ ] **COMMITS**: Formato semântico obrigatório

## 🚫 VIOLAÇÕES CRÍTICAS

| Violação | Transformação |
|----------|---------------|
| `func (s *Server)` | `func (server *Server)` |
| `func (s *Server) ping()` sem usar `s.*` | `func ping()` em utils.go |
| `if err != nil` | `if hasError(err)` |
| `if len(x) == 0` | `if isEmpty(x)` |
| `if/else` | return early |
| `switch` | map dispatch |
| `interface{}` | `any` |
| Função independente fora utils.go | Mover para utils.go |
| Qualquer comentário | Remover, usar nomes descritivos |
| Commit não semântico | `tipo: descrição imperativo minúsculo` |

## 🔄 PROCESSO DE TRABALHO

1. **SEMPRE** perguntar antes de executar
2. Identificar violações dos padrões
3. Propor correções específicas
4. Aguardar confirmação do usuário
5. Executar apenas após confirmação
6. Verificar conformidade com checklist
7. Executar testes para validar
8. **Commit semântico**: Usar formato `tipo: descrição`