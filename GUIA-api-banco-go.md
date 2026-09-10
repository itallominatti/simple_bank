# Guia completo: API REST de banco com Go + Gin, usando DDD, Clean Architecture e Docker Compose

Este guia parte do zero. Ele não assume que você conhece Go, Gin, banco de dados, Docker, DDD ou Clean Architecture. Primeiro você vai entender os conceitos (Parte 1) e depois construir o projeto arquivo por arquivo (Parte 2). Todo o código deste guia foi compilado e testado.

Dica de leitura: não tente decorar tudo na primeira passada. Leia a Parte 1 com calma, construa o projeto na Parte 2 e depois volte à Parte 1. Ela vai fazer muito mais sentido quando você tiver o código na sua frente.

---

## O que vamos construir

Uma API de um banco simples, onde é possível:

| Ação | Método HTTP | Endereço (rota) |
|---|---|---|
| Verificar se a API está no ar | `GET` | `/health` |
| Abrir uma conta | `POST` | `/accounts` |
| Consultar uma conta (saldo) | `GET` | `/accounts/{id}` |
| Depositar | `POST` | `/accounts/{id}/deposit` |
| Sacar | `POST` | `/accounts/{id}/withdraw` |
| Transferir entre contas | `POST` | `/transfers` |
| Ver o extrato | `GET` | `/accounts/{id}/transactions` |

Tecnologias: **Go** (linguagem), **Gin** (framework web), **PostgreSQL** (banco de dados), **Docker Compose** (para rodar tudo junto) e **Adminer** (uma tela web para você enxergar o banco de dados).

---

# PARTE 1 — Os conceitos, explicados do jeito mais simples possível

## 1.1 O que é uma API REST?

Pense em um **restaurante**. Você (o cliente) não entra na cozinha; você faz o pedido ao **garçom**, que leva até a cozinha e depois traz o prato. A **API** é o garçom: um programa que recebe pedidos de outros programas (um app de celular, um site, o terminal) e devolve respostas.

**REST** é um conjunto de convenções para organizar esses pedidos pela internet, usando o protocolo **HTTP** (o mesmo que o navegador usa). Um pedido HTTP tem:

- **Método** (o verbo, o que você quer fazer): `GET` = buscar, `POST` = criar/executar uma ação, `PUT`/`PATCH` = alterar, `DELETE` = apagar.
- **Rota** (o endereço do recurso): `/accounts/123`.
- **Corpo** (opcional, os dados enviados), normalmente em **JSON**.

**JSON** é só um formato de texto para representar dados, fácil para humanos e máquinas lerem:

```json
{ "owner_name": "Maria Silva", "amount": 10000 }
```

A resposta sempre vem com um **código de status**, um número que resume o resultado:

| Código | Significado | Quando usamos neste projeto |
|---|---|---|
| 200 | OK | Deu tudo certo |
| 201 | Created | Uma conta foi criada |
| 400 | Bad Request | O pedido veio errado (ex.: valor negativo) |
| 404 | Not Found | A conta não existe |
| 422 | Unprocessable Entity | O pedido é válido, mas a regra de negócio impede (ex.: saldo insuficiente) |
| 500 | Internal Server Error | Um erro inesperado no servidor |

## 1.2 Go em poucos minutos

Go é uma linguagem criada pelo Google, conhecida por ser simples, rápida e muito usada em APIs. Você só precisa destes conceitos para acompanhar o guia:

```go
package main // todo arquivo pertence a um "pacote" (uma pasta = um pacote)

import "fmt" // importa código de outro pacote

// struct = um "molde" que agrupa dados (parecido com uma ficha cadastral)
type Pessoa struct {
    Nome  string // texto
    Idade int    // número inteiro
}

// método = uma função que "pertence" a uma struct
func (p Pessoa) Saudacao() string {
    return "Olá, " + p.Nome
}

// função que pode dar erro: em Go, o erro é devolvido como um valor comum
func Dividir(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("não dá para dividir por zero")
    }
    return a / b, nil // nil = "nada", ou seja, sem erro
}

func main() {
    p := Pessoa{Nome: "Ana", Idade: 30} // := cria uma variável
    fmt.Println(p.Saudacao())

    resultado, err := Dividir(10, 0)
    if err != nil { // este "if err != nil" vai aparecer o tempo todo
        fmt.Println("erro:", err)
        return
    }
    fmt.Println(resultado)
}
```

Mais quatro ideias importantes:

**Maiúscula = público, minúscula = privado.** Um nome que começa com letra maiúscula (`Deposit`, `Account`) pode ser usado por outros pacotes. Com minúscula (`balance`), só pode ser usado dentro do próprio pacote. Vamos usar isso para proteger o saldo da conta.

**Interface = um contrato.** Uma interface diz "quem quiser ser X precisa ter estes métodos". Pense numa **tomada**: ela não sabe se vai ligar uma geladeira ou um carregador; ela só exige que o plugue tenha o formato certo. Qualquer struct que tenha os métodos da interface "encaixa" nela automaticamente.

**Ponteiro (`*`).** `*Account` significa "o endereço de uma conta", e não uma cópia dela. Usamos isso para que, quando a conta for alterada, todos que a estão usando vejam a alteração. Por enquanto, basta saber isso.

**`context.Context`.** Um parâmetro que aparece como primeiro argumento em muitas funções. Ele carrega informações da requisição, como "o cliente desistiu, pode parar". Você só precisa repassá-lo adiante.

## 1.3 O que é o Gin?

Go já sabe criar servidores web sozinho, mas o **Gin** é um *framework* (uma caixa de ferramentas) que deixa isso mais fácil. Ele cuida de:

- **Rotas:** "quando chegar um `POST /accounts`, chame a função `CreateAccount`";
- **Ler o JSON** que chega e transformá-lo numa struct do Go;
- **Responder em JSON** com o código de status certo;
- **Logs** de cada requisição e proteção contra quedas.

## 1.4 O que é um banco de dados (e o PostgreSQL)?

Se o programa guardasse as contas só na memória, tudo sumiria ao desligá-lo. O **banco de dados** é onde as informações ficam guardadas de forma permanente e organizada.

O **PostgreSQL** é um banco de dados **relacional**: ele guarda os dados em **tabelas**, que se parecem muito com planilhas do Excel.

Tabela `accounts` (contas):

| id | owner_name | balance | created_at |
|---|---|---|---|
| 9a67b859-... | Maria Silva | 5950 | 2026-09-10 ... |
| 76811e08-... | João Souza | 1500 | 2026-09-10 ... |

Alguns termos:

- **Coluna:** cada "campo" (id, nome, saldo).
- **Linha (registro):** cada conta.
- **Chave primária (primary key):** a coluna que identifica cada linha de forma única (o `id`). Não pode repetir.
- **Chave estrangeira (foreign key):** uma coluna que aponta para a chave primária de outra tabela. Na tabela `transactions`, a coluna `account_id` aponta para uma conta que precisa existir.
- **SQL:** a linguagem para conversar com o banco. Exemplos: `SELECT` (buscar), `INSERT` (inserir), `UPDATE` (alterar).
- **UUID:** um identificador aleatório e gigante (ex.: `9a67b859-8c59-4561-84c8-da8a6eec82ad`), praticamente impossível de repetir.
- **Transação:** um "pacote" de comandos que é salvo por inteiro ou descartado por inteiro. Em uma transferência, tirar o dinheiro de A e colocar em B precisam acontecer juntos. Se algo falhar no meio, o banco desfaz tudo (isso se chama *rollback*). Se der certo, confirma tudo (*commit*).

## 1.5 O que é Docker e Docker Compose?

Sabe o clássico "na minha máquina funciona"? O **Docker** resolve isso empacotando o programa junto com tudo o que ele precisa para rodar.

- **Imagem:** uma "receita" (ou forma de bolo) pronta e imutável. Ex.: a imagem oficial do PostgreSQL.
- **Container:** a receita "em execução" (o bolo pronto). Você pode criar vários containers a partir da mesma imagem. Ele funciona como um mini computador isolado dentro do seu.
- **Dockerfile:** o arquivo onde **você escreve a receita** da sua própria imagem (no nosso caso, "pegue o Go, compile a API, gere uma imagem pequena com o executável").
- **Volume:** um "HD virtual". Containers são descartáveis; o que está dentro deles some quando são apagados. O volume guarda os dados do banco fora do container.
- **Docker Compose:** um arquivo (`docker-compose.yml`) que descreve **vários containers de uma vez** e como eles se conectam. Com um único comando, `docker compose up`, sobe o banco, a API e o Adminer.

Com isso, você **não precisa instalar PostgreSQL nem Go** no seu computador. Só o Docker.

## 1.6 O que é DDD (Domain-Driven Design)?

**DDD** ("Design Orientado ao Domínio") é uma forma de pensar o software em que **o mais importante é o negócio**, e não a tecnologia. O "domínio" é o assunto do sistema, que aqui é "banco". A ideia central: as regras do negócio ("não pode sacar mais que o saldo") ficam num lugar central, bem protegido e escrito na linguagem de quem entende do negócio.

Os conceitos de DDD que usaremos:

| Conceito | Explicação simples | No nosso projeto |
|---|---|---|
| **Linguagem ubíqua** | Programadores e especialistas do negócio usam as mesmas palavras. Se o gerente fala "sacar", o código tem um método `Withdraw` (sacar). | `Deposit`, `Withdraw`, `Transfer` |
| **Entidade** | Algo que tem **identidade própria** e muda com o tempo. Duas contas com o mesmo saldo continuam sendo contas diferentes, porque têm IDs diferentes. | `Account` |
| **Objeto de Valor** | Algo definido **só pelo seu valor**, sem identidade. Uma nota de R$ 10 vale o mesmo que qualquer outra nota de R$ 10. | `Money` |
| **Agregado** | Um grupo de objetos tratados como uma unidade, com uma "porta de entrada" (a raiz). Ninguém mexe no saldo por fora; é preciso pedir à conta. | `Account` é a raiz |
| **Repositório** | O "arquivo" onde as entidades são guardadas e buscadas. O domínio só define **o que** o repositório faz, não **como**. | `AccountRepository` |
| **Serviço de Domínio** | Uma regra de negócio que envolve mais de uma entidade e, por isso, não cabe numa só. | `Transfer(de, para, valor)` |

## 1.7 O que é Clean Architecture (Arquitetura Limpa)?

É uma forma de **organizar o código em camadas**, em que cada camada tem uma responsabilidade. Voltando ao restaurante:

- **Garçom (camada de entrega / handler):** recebe o pedido do cliente e entrega a resposta. Não cozinha.
- **Chef (camada de aplicação / casos de uso):** coordena o preparo de cada prato, seguindo as receitas.
- **Livro de receitas (camada de domínio):** as regras. "Bolo leva 3 ovos". Não importa se o forno é a gás ou elétrico.
- **Despensa e fornecedores (camada de infraestrutura):** onde os ingredientes são guardados e buscados (o banco de dados).

A **Regra de Ouro (Regra da Dependência)**: as camadas de fora podem conhecer as de dentro, mas **as de dentro nunca conhecem as de fora**.

```
   handler (HTTP / Gin)            infrastructure (PostgreSQL)
            \                              /
             \   (conhecem / dependem de) /
              ▼                          ▼
             application (casos de uso)
                         │
                         ▼
               domain (regras do banco)   ← não conhece NINGUÉM
```

O domínio não sabe que existe Gin, JSON, HTTP, PostgreSQL ou Docker. Por que isso é bom?

1. **Testar é fácil:** dá para testar as regras de negócio sem banco de dados e sem servidor (você fará isso no final).
2. **Trocar é fácil:** se amanhã você trocar PostgreSQL por MySQL, ou Gin por outro framework, as regras de negócio não mudam uma vírgula.
3. **Entender é fácil:** cada coisa tem seu lugar. Precisa mudar uma regra de saque? Vá direto no domínio.

**"Mas o repositório com SQL fica na infraestrutura, e o caso de uso precisa dele... isso não quebra a regra?"** Ótima pergunta. Aqui entra o truque chamado **Inversão de Dependência**: o domínio define a **interface** (a tomada: "preciso de alguém que busque e salve contas") e a infraestrutura cria o **plugue** que encaixa nela (a implementação com PostgreSQL). O caso de uso só conhece a tomada. Quem liga o plugue na tomada é o `main.go`, na hora de iniciar o programa.

## 1.8 Como o projeto fica organizado

```
banco-api/
├── cmd/
│   └── api/
│       └── main.go                  ← ponto de partida: monta e liga todas as peças
├── internal/
│   ├── domain/                      ← CAMADA DE DOMÍNIO (regras do negócio / DDD)
│   │   ├── errors.go                   erros de negócio
│   │   ├── money.go                    objeto de valor Money
│   │   ├── account.go                  entidade Account
│   │   ├── transaction.go              movimentação (linha do extrato)
│   │   ├── transfer.go                 serviço de domínio de transferência
│   │   ├── repository.go               contratos (interfaces) dos repositórios
│   │   └── account_test.go             testes das regras
│   ├── application/                 ← CAMADA DE APLICAÇÃO (casos de uso)
│   │   ├── unit_of_work.go             contrato de "tudo ou nada"
│   │   └── account_service.go          abrir conta, depositar, sacar, transferir...
│   ├── infrastructure/
│   │   └── postgres/                ← CAMADA DE INFRAESTRUTURA (banco de dados)
│   │       ├── db.go                   conexão com o PostgreSQL
│   │       ├── account_repository.go   SQL das contas
│   │       ├── transaction_repository.go SQL do extrato
│   │       └── unit_of_work.go         transação do banco
│   └── handler/                     ← CAMADA DE ENTREGA (HTTP com Gin)
│       ├── dto.go                      formatos do JSON de entrada e saída
│       ├── errors.go                   erro de negócio → código HTTP
│       ├── account_handler.go          funções que atendem cada rota
│       └── router.go                   tabela de rotas
├── migrations/
│   └── 001_create_tables.sql        ← criação das tabelas no banco
├── Dockerfile                       ← receita da imagem da API
├── docker-compose.yml               ← orquestra banco + API + Adminer
├── .dockerignore
├── requests.http                    ← requisições prontas para testar no VS Code
└── go.mod / go.sum                  ← lista de dependências (gerados por comando)
```

A pasta `internal` é especial em Go: o código dentro dela só pode ser importado por este projeto. A pasta `cmd` guarda os programas executáveis.

## 1.9 O caminho de uma requisição

Veja o que acontece quando alguém faz um saque:

```
Cliente (curl, app, site)
   │  POST /accounts/9a67.../withdraw     corpo: {"amount": 500}
   ▼
router.go ─────────── "essa rota é da função Withdraw"
   ▼
account_handler.go ── lê o JSON, chama o caso de uso
   ▼
account_service.go ── abre uma transação (UnitOfWork)
   │                  1. busca a conta no repositório ──────► PostgreSQL
   │                  2. pede para a conta sacar ──► account.go (verifica o saldo!)
   │                  3. salva o novo saldo ─────────────────► PostgreSQL
   │                  4. registra no extrato ────────────────► PostgreSQL
   │                  5. confirma a transação (commit)
   ▼
account_handler.go ── transforma o resultado em JSON (ou o erro em 422, 404...)
   ▼
Cliente recebe:  200 {"id": "9a67...", "balance": 9500, "balance_formatted": "R$ 95,00", ...}
```

---

# PARTE 2 — Mão na massa

## Etapa 1 — Instalar as ferramentas

**1. Docker Desktop** (obrigatório). Baixe em https://www.docker.com/products/docker-desktop/ e instale. No Windows, o instalador vai pedir para ativar o **WSL 2**; aceite. No Linux, instale o Docker Engine e o plugin do Compose seguindo a documentação oficial.

Depois de instalar, **abra o Docker Desktop** (ele precisa estar rodando) e confira no terminal:

```bash
docker --version
docker compose version
```

Se os dois comandos mostrarem uma versão, está pronto.

> **O que é o terminal?** É a janela onde você digita comandos. No Mac: app "Terminal". No Windows: "PowerShell" ou, melhor ainda, o "Git Bash" (vem com o Git, https://git-scm.com). No Linux: "Terminal".

**2. VS Code** (recomendado): https://code.visualstudio.com. Instale estas extensões (ícone de quadradinhos na barra lateral):

- **Go** (da equipe Go do Google): colore o código e aponta erros;
- **REST Client** (de Huachao Mao): para testar a API clicando num botão.

**3. Go** (opcional). Com Docker você não precisa do Go instalado, porque o Docker compila tudo. Mas, se quiser que o VS Code aponte erros enquanto você digita, instale em https://go.dev/dl/.

## Etapa 2 — Criar as pastas do projeto

No terminal, vá até onde quer guardar o projeto (por exemplo, a pasta Documentos) e rode:

**Mac, Linux ou Git Bash:**

```bash
mkdir banco-api
cd banco-api
mkdir -p cmd/api internal/domain internal/application internal/infrastructure/postgres internal/handler migrations
```

**Windows PowerShell:**

```powershell
mkdir banco-api
cd banco-api
mkdir cmd/api, internal/domain, internal/application, internal/infrastructure/postgres, internal/handler, migrations
```

Agora abra a pasta no VS Code: `code .` (ou pelo menu *File > Open Folder*).

## Etapa 3 — Criar o módulo Go (`go.mod`)

Todo projeto Go tem um arquivo `go.mod`, que diz o **nome do projeto** e **quais bibliotecas ele usa** (como a lista de ingredientes de uma receita). Vamos criá-lo usando um container temporário com Go (assim você não precisa instalá-lo):

**Mac, Linux ou Git Bash:**

```bash
docker run --rm -v "$(pwd)":/app -w /app golang:1.27-alpine go mod init banco-api
```

**Windows PowerShell:**

```powershell
docker run --rm -v "${PWD}:/app" -w /app golang:1.27-alpine go mod init banco-api
```

Entendendo o comando:

| Parte | Significado |
|---|---|
| `docker run` | cria e roda um container |
| `--rm` | apaga o container quando o comando terminar |
| `-v "$(pwd)":/app` | "espelha" a pasta atual do seu computador dentro do container, em `/app` |
| `-w /app` | o comando roda dentro da pasta `/app` |
| `golang:1.27-alpine` | a imagem oficial do Go (versão 1.27, numa base Linux pequena chamada Alpine) |
| `go mod init banco-api` | o comando do Go que cria o `go.mod` com o nome `banco-api` |

Na primeira vez, o Docker baixa a imagem (demora um pouco). Depois disso, deve aparecer um arquivo `go.mod` na sua pasta. O nome `banco-api` é o que usaremos nos `import` (ex.: `banco-api/internal/domain`).

(Se você instalou o Go, o equivalente seria só `go mod init banco-api`.)

## Etapa 4 — A camada de DOMÍNIO (o coração do sistema)

Começamos pelo domínio de propósito: ele é o centro e não depende de nada. Crie cada arquivo abaixo dentro de `internal/domain/`.

### 4.1 `internal/domain/errors.go` — os erros de negócio

```go
package domain

import "errors"

// Erros de negócio. Eles descrevem O QUE deu errado na linguagem do banco,
// sem falar nada de HTTP ou de banco de dados.
var (
	ErrAccountNotFound     = errors.New("conta não encontrada")
	ErrInvalidOwnerName    = errors.New("o nome do titular é obrigatório")
	ErrInvalidAmount       = errors.New("o valor deve ser maior que zero")
	ErrInsufficientFunds   = errors.New("saldo insuficiente")
	ErrSameAccountTransfer = errors.New("não é possível transferir para a mesma conta")
)
```

**O que está acontecendo:** criamos "erros com nome". Quando algo der errado, o domínio devolve um desses erros e as outras camadas conseguem descobrir exatamente qual foi (usando `errors.Is`). Repare que as mensagens falam de negócio ("saldo insuficiente"), nunca de tecnologia ("erro 422" ou "falha no SQL").

### 4.2 `internal/domain/money.go` — o Objeto de Valor "dinheiro"

```go
package domain

import "fmt"

// Money é um OBJETO DE VALOR: representa uma quantia de dinheiro.
// Guardamos sempre em CENTAVOS, usando número inteiro.
// Exemplo: R$ 10,50 é guardado como 1050.
//
// Por que não usar float (número com vírgula)? Porque o computador não
// consegue representar alguns decimais com exatidão (0.1 + 0.2 = 0.30000000000000004).
// Em banco, um centavo errado é um problema sério.
type Money int64

// IsPositive diz se o valor é maior que zero.
func (m Money) IsPositive() bool {
	return m > 0
}

// String formata o valor para exibição. Ex.: 1050 -> "R$ 10,50".
func (m Money) String() string {
	value := int64(m)
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	return fmt.Sprintf("%sR$ %d,%02d", sign, value/100, value%100)
}
```

**O que está acontecendo:** `type Money int64` cria um tipo novo baseado num número inteiro de 64 bits. Poderíamos usar `int64` diretamente, mas dar um nome ao conceito ("isto é dinheiro, em centavos") deixa o código mais claro e permite colocar comportamento nele, como formatar para `R$ 10,50`. Esse é um **Objeto de Valor**: não tem ID; 1050 centavos é sempre igual a 1050 centavos.

A regra "**dinheiro sempre em centavos, nunca com vírgula**" é uma das mais importantes em qualquer sistema financeiro.

### 4.3 `internal/domain/account.go` — a Entidade "conta"

```go
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Account é uma ENTIDADE (e a raiz do AGREGADO "Conta").
// Entidade = algo que tem identidade própria (o ID) e cujo estado muda com o tempo.
//
// Repare que os campos começam com letra minúscula: em Go isso significa
// que eles são PRIVADOS, ou seja, código de fora deste pacote não consegue
// fazer "conta.balance = 1000000". A única forma de mudar o saldo é pelos
// métodos Deposit e Withdraw, que aplicam as regras do negócio.
type Account struct {
	id        string
	ownerName string
	balance   Money
	createdAt time.Time
	updatedAt time.Time
}

// NewAccount cria uma conta NOVA, validando as regras de criação.
func NewAccount(ownerName string) (*Account, error) {
	ownerName = strings.TrimSpace(ownerName)
	if ownerName == "" {
		return nil, ErrInvalidOwnerName
	}

	now := time.Now().UTC()
	return &Account{
		id:        uuid.NewString(),
		ownerName: ownerName,
		balance:   0,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// RestoreAccount recria uma conta que JÁ EXISTE (por exemplo, lida do banco de dados).
// Não valida nada, porque os dados já foram validados quando a conta foi criada.
// Deve ser usada apenas pelos repositórios.
func RestoreAccount(id, ownerName string, balance Money, createdAt, updatedAt time.Time) *Account {
	return &Account{
		id:        id,
		ownerName: ownerName,
		balance:   balance,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Deposit adiciona dinheiro na conta.
// Regra de negócio: só é possível depositar valores positivos.
func (a *Account) Deposit(amount Money) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	a.balance += amount
	a.updatedAt = time.Now().UTC()
	return nil
}

// Withdraw retira dinheiro da conta.
// Regras de negócio: o valor deve ser positivo e a conta precisa ter saldo suficiente.
func (a *Account) Withdraw(amount Money) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	if a.balance < amount {
		return ErrInsufficientFunds
	}
	a.balance -= amount
	a.updatedAt = time.Now().UTC()
	return nil
}

// "Getters": permitem LER os dados, mas não alterá-los.

func (a *Account) ID() string           { return a.id }
func (a *Account) OwnerName() string    { return a.ownerName }
func (a *Account) Balance() Money       { return a.balance }
func (a *Account) CreatedAt() time.Time { return a.createdAt }
func (a *Account) UpdatedAt() time.Time { return a.updatedAt }
```

**O que está acontecendo:**

- A struct `Account` tem campos **privados** (minúsculos). Isso é o que o DDD chama de **encapsulamento**: a conta protege seus próprios dados. É impossível alguém de fora escrever `conta.balance = -500`.
- `NewAccount` é a **única porta para criar uma conta nova**, e ela valida o nome. Uma conta inválida nunca chega a existir.
- `RestoreAccount` serve para "reconstruir" uma conta que veio do banco de dados. É diferente de criar uma conta nova (não gera um novo ID, não zera o saldo).
- `Deposit` e `Withdraw` são as **regras de negócio**. Repare que elas não sabem nada de banco de dados ou HTTP; são só lógica pura.
- Os métodos no final (`ID()`, `Balance()`...) permitem **ler** os dados, mas não alterá-los.
- `uuid.NewString()` usa uma biblioteca externa (`github.com/google/uuid`) para gerar o ID. Ela será baixada na Etapa 9.

### 4.4 `internal/domain/transaction.go` — a movimentação (linha do extrato)

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// TransactionType é o tipo de movimentação que aconteceu na conta.
type TransactionType string

const (
	TransactionDeposit     TransactionType = "DEPOSIT"      // depósito
	TransactionWithdraw    TransactionType = "WITHDRAW"     // saque
	TransactionTransferIn  TransactionType = "TRANSFER_IN"  // transferência recebida
	TransactionTransferOut TransactionType = "TRANSFER_OUT" // transferência enviada
)

// Transaction é uma movimentação bancária (uma linha do extrato).
// Ela é um registro histórico: depois de criada, nunca muda.
// (Não confunda com "transação de banco de dados", que aparece mais adiante
// com o nome UnitOfWork.)
type Transaction struct {
	ID        string
	AccountID string
	Type      TransactionType
	Amount    Money
	CreatedAt time.Time
}

// NewTransaction cria um novo registro de movimentação.
func NewTransaction(accountID string, txType TransactionType, amount Money) *Transaction {
	return &Transaction{
		ID:        uuid.NewString(),
		AccountID: accountID,
		Type:      txType,
		Amount:    amount,
		CreatedAt: time.Now().UTC(),
	}
}
```

**O que está acontecendo:** cada depósito, saque ou transferência gera um registro para o extrato. `TransactionType` é um texto com valores fixos (as constantes `const`), para evitar erros de digitação espalhados pelo código. Aqui os campos são públicos porque a movimentação é só um registro histórico que nunca muda depois de criado.

### 4.5 `internal/domain/transfer.go` — o Serviço de Domínio

```go
package domain

// Transfer é um SERVIÇO DE DOMÍNIO: uma regra de negócio que envolve
// MAIS DE UMA entidade e, por isso, não pertence a uma conta só.
//
// Ela move dinheiro de uma conta para outra, garantindo as regras:
//   - não pode transferir para a mesma conta;
//   - o valor deve ser positivo;
//   - a conta de origem precisa ter saldo.
func Transfer(from, to *Account, amount Money) error {
	if from.ID() == to.ID() {
		return ErrSameAccountTransfer
	}

	// Primeiro sacamos. Se falhar (ex.: sem saldo), nada foi alterado.
	if err := from.Withdraw(amount); err != nil {
		return err
	}

	// Depois depositamos. Aqui não tem como falhar, pois o valor já foi validado no saque.
	return to.Deposit(amount)
}
```

**O que está acontecendo:** transferir envolve **duas** contas, então a regra não pertence a uma conta só. Em DDD, isso vira um **Serviço de Domínio**: uma função que coordena entidades seguindo as regras do negócio. A ordem importa: sacamos primeiro porque, se não houver saldo, nada é alterado.

### 4.6 `internal/domain/repository.go` — os contratos dos repositórios

```go
package domain

import "context"

// Aqui ficam os CONTRATOS (interfaces) dos repositórios.
//
// Um repositório é quem sabe guardar e buscar entidades. O domínio só diz
// "eu preciso de alguém que faça isso"; ele NÃO sabe se os dados vão para
// PostgreSQL, MySQL, um arquivo ou a memória. Quem implementa isso é a
// camada de infraestrutura.
//
// O "context.Context" que aparece em todo lugar é um padrão do Go: ele
// carrega informações da requisição, como "o cliente desistiu/cancelou".

// AccountRepository guarda e busca contas.
type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	FindByID(ctx context.Context, id string) (*Account, error)
	// FindByIDForUpdate busca a conta e a "tranca" até o fim da transação,
	// para que duas operações não alterem o mesmo saldo ao mesmo tempo.
	FindByIDForUpdate(ctx context.Context, id string) (*Account, error)
	Update(ctx context.Context, account *Account) error
}

// TransactionRepository guarda e busca movimentações (extrato).
type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	ListByAccountID(ctx context.Context, accountID string) ([]*Transaction, error)
}
```

**O que está acontecendo:** estas são as **tomadas** da analogia. O domínio declara: "preciso de alguém que saiba criar, buscar e atualizar contas". Não há uma linha de SQL aqui. A implementação real (o plugue) virá na infraestrutura.

O método `FindByIDForUpdate` merece atenção. Imagine duas pessoas sacando R$ 100 ao mesmo tempo de uma conta com R$ 100. Sem proteção, as duas leriam "saldo = 100", as duas achariam que podem sacar e a conta terminaria com saldo errado. O "for update" tranca a conta: a segunda operação espera a primeira terminar e, quando for a vez dela, já vai ler o saldo atualizado.

### 4.7 `internal/domain/account_test.go` — testando as regras

```go
package domain

import (
	"errors"
	"testing"
)

// Estes testes rodam SEM banco de dados, SEM Docker e SEM HTTP.
// Isso só é possível porque as regras de negócio estão isoladas no domínio.

func TestNewAccount_RequiresOwnerName(t *testing.T) {
	_, err := NewAccount("   ")
	if !errors.Is(err, ErrInvalidOwnerName) {
		t.Fatalf("esperava ErrInvalidOwnerName, recebi %v", err)
	}
}

func TestDepositAndWithdraw(t *testing.T) {
	account, _ := NewAccount("Maria")

	if err := account.Deposit(1000); err != nil { // R$ 10,00
		t.Fatalf("depósito falhou: %v", err)
	}
	if err := account.Withdraw(300); err != nil { // R$ 3,00
		t.Fatalf("saque falhou: %v", err)
	}
	if account.Balance() != 700 {
		t.Fatalf("saldo esperado 700, recebi %d", account.Balance())
	}
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	account, _ := NewAccount("João")

	err := account.Withdraw(100)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("esperava ErrInsufficientFunds, recebi %v", err)
	}
	if account.Balance() != 0 {
		t.Fatalf("o saldo não deveria ter mudado")
	}
}

func TestTransfer(t *testing.T) {
	from, _ := NewAccount("Ana")
	to, _ := NewAccount("Beto")
	_ = from.Deposit(5000)

	if err := Transfer(from, to, 2000); err != nil {
		t.Fatalf("transferência falhou: %v", err)
	}
	if from.Balance() != 3000 || to.Balance() != 2000 {
		t.Fatalf("saldos errados: origem=%d destino=%d", from.Balance(), to.Balance())
	}
}

func TestTransfer_SameAccount(t *testing.T) {
	account, _ := NewAccount("Ana")
	_ = account.Deposit(5000)

	err := Transfer(account, account, 100)
	if !errors.Is(err, ErrSameAccountTransfer) {
		t.Fatalf("esperava ErrSameAccountTransfer, recebi %v", err)
	}
}

func TestMoneyString(t *testing.T) {
	if got := Money(1050).String(); got != "R$ 10,50" {
		t.Fatalf("esperava R$ 10,50, recebi %s", got)
	}
}
```

**O que está acontecendo:** em Go, arquivos terminados em `_test.go` são testes automáticos. Cada função `TestXxx` verifica uma regra. Vamos rodá-los na Etapa 15. O ponto principal: **estes testes não precisam de banco de dados, Docker ou internet**, porque o domínio é independente. Esse é um dos maiores benefícios da Clean Architecture.

## Etapa 5 — A camada de APLICAÇÃO (casos de uso)

Crie os arquivos dentro de `internal/application/`.

### 5.1 `internal/application/unit_of_work.go` — o contrato "tudo ou nada"

```go
package application

import (
	"context"

	"banco-api/internal/domain"
)

// Repositories agrupa os repositórios que podem ser usados dentro de uma UnitOfWork.
type Repositories struct {
	Accounts     domain.AccountRepository
	Transactions domain.TransactionRepository
}

// UnitOfWork ("unidade de trabalho") garante que várias operações aconteçam
// como se fossem UMA SÓ: ou tudo dá certo, ou nada é salvo.
//
// Exemplo: numa transferência, precisamos (1) tirar da conta A, (2) colocar
// na conta B e (3) registrar o extrato. Se o passo 2 falhar, o passo 1 NÃO
// pode ficar salvo, senão o dinheiro "some".
//
// Tudo que for feito com os repositórios recebidos em "fn" faz parte do mesmo pacote.
// Se "fn" retornar erro, tudo é desfeito.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(repos Repositories) error) error
}
```

**O que está acontecendo:** esta é outra "tomada". O caso de uso diz: "preciso executar várias operações como se fossem uma só". A implementação com transação do PostgreSQL virá na infraestrutura. A função `Do` recebe outra função (`fn`) com o trabalho a ser feito. Se `fn` der erro, tudo é desfeito.

> Não confunda **Transaction** (a movimentação bancária do domínio, a linha do extrato) com **transação de banco de dados** (o "pacote tudo ou nada"). Por isso chamamos a segunda de **UnitOfWork**.

### 5.2 `internal/application/account_service.go` — os casos de uso

```go
package application

import (
	"context"

	"banco-api/internal/domain"
)

// AccountService contém os CASOS DE USO do sistema: as ações que um usuário
// pode realizar (abrir conta, depositar, sacar, transferir, ver extrato).
//
// Ele funciona como um maestro: busca os dados nos repositórios, pede para
// o domínio aplicar as regras e manda salvar o resultado.
// Ele NÃO sabe nada de HTTP, JSON, Gin ou PostgreSQL.
type AccountService struct {
	accounts     domain.AccountRepository
	transactions domain.TransactionRepository
	uow          UnitOfWork
}

// NewAccountService recebe suas dependências "de fora" (isso se chama
// INJEÇÃO DE DEPENDÊNCIA). Assim, em testes, podemos passar versões falsas.
func NewAccountService(
	accounts domain.AccountRepository,
	transactions domain.TransactionRepository,
	uow UnitOfWork,
) *AccountService {
	return &AccountService{
		accounts:     accounts,
		transactions: transactions,
		uow:          uow,
	}
}

// CreateAccount abre uma nova conta.
func (s *AccountService) CreateAccount(ctx context.Context, ownerName string) (*domain.Account, error) {
	account, err := domain.NewAccount(ownerName)
	if err != nil {
		return nil, err
	}

	if err := s.accounts.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

// GetAccount busca uma conta pelo ID.
func (s *AccountService) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	return s.accounts.FindByID(ctx, id)
}

// Deposit deposita um valor na conta e registra no extrato.
func (s *AccountService) Deposit(ctx context.Context, accountID string, amount domain.Money) (*domain.Account, error) {
	var account *domain.Account

	err := s.uow.Do(ctx, func(repos Repositories) error {
		var err error
		account, err = repos.Accounts.FindByIDForUpdate(ctx, accountID)
		if err != nil {
			return err
		}

		if err := account.Deposit(amount); err != nil {
			return err
		}

		if err := repos.Accounts.Update(ctx, account); err != nil {
			return err
		}

		return repos.Transactions.Create(ctx,
			domain.NewTransaction(account.ID(), domain.TransactionDeposit, amount))
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

// Withdraw saca um valor da conta e registra no extrato.
func (s *AccountService) Withdraw(ctx context.Context, accountID string, amount domain.Money) (*domain.Account, error) {
	var account *domain.Account

	err := s.uow.Do(ctx, func(repos Repositories) error {
		var err error
		account, err = repos.Accounts.FindByIDForUpdate(ctx, accountID)
		if err != nil {
			return err
		}

		if err := account.Withdraw(amount); err != nil {
			return err
		}

		if err := repos.Accounts.Update(ctx, account); err != nil {
			return err
		}

		return repos.Transactions.Create(ctx,
			domain.NewTransaction(account.ID(), domain.TransactionWithdraw, amount))
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

// Transfer move dinheiro de uma conta para outra.
func (s *AccountService) Transfer(ctx context.Context, fromID, toID string, amount domain.Money) error {
	// Validações rápidas antes de ir ao banco de dados.
	if fromID == toID {
		return domain.ErrSameAccountTransfer
	}
	if !amount.IsPositive() {
		return domain.ErrInvalidAmount
	}

	return s.uow.Do(ctx, func(repos Repositories) error {
		// Trancamos as duas contas SEMPRE na mesma ordem (menor ID primeiro).
		// Se duas transferências A->B e B->A acontecerem juntas e cada uma
		// trancar uma conta primeiro, uma ficaria esperando a outra para sempre
		// (isso se chama "deadlock"). Ordenar evita esse problema.
		firstID, secondID := fromID, toID
		if secondID < firstID {
			firstID, secondID = secondID, firstID
		}

		first, err := repos.Accounts.FindByIDForUpdate(ctx, firstID)
		if err != nil {
			return err
		}
		second, err := repos.Accounts.FindByIDForUpdate(ctx, secondID)
		if err != nil {
			return err
		}

		// Descobre quem é a origem e quem é o destino.
		from, to := first, second
		if from.ID() != fromID {
			from, to = second, first
		}

		// A regra de negócio mora no domínio.
		if err := domain.Transfer(from, to, amount); err != nil {
			return err
		}

		// Salva os dois saldos e registra as duas linhas de extrato.
		if err := repos.Accounts.Update(ctx, from); err != nil {
			return err
		}
		if err := repos.Accounts.Update(ctx, to); err != nil {
			return err
		}
		if err := repos.Transactions.Create(ctx,
			domain.NewTransaction(from.ID(), domain.TransactionTransferOut, amount)); err != nil {
			return err
		}
		return repos.Transactions.Create(ctx,
			domain.NewTransaction(to.ID(), domain.TransactionTransferIn, amount))
	})
}

// ListTransactions retorna o extrato de uma conta.
func (s *AccountService) ListTransactions(ctx context.Context, accountID string) ([]*domain.Transaction, error) {
	// Garante que a conta existe (senão retornaríamos uma lista vazia sem avisar).
	if _, err := s.accounts.FindByID(ctx, accountID); err != nil {
		return nil, err
	}
	return s.transactions.ListByAccountID(ctx, accountID)
}
```

**O que está acontecendo:** cada método é um **caso de uso**, ou seja, uma ação que o usuário pode fazer. Todos seguem o mesmo roteiro do "chef":

1. buscar os dados (pelos repositórios);
2. pedir ao domínio para aplicar a regra (`account.Deposit(...)`);
3. salvar o resultado e registrar no extrato.

Note como `Deposit`, `Withdraw` e `Transfer` rodam **dentro** de `s.uow.Do(...)`. Assim, se o saldo for atualizado mas o registro do extrato falhar, o saldo volta ao que era. Nunca fica "meio salvo".

Na transferência há um cuidado extra: trancamos as contas sempre na mesma ordem. Se a transferência "A→B" trancasse A e esperasse B, enquanto outra "B→A" trancasse B e esperasse A, as duas ficariam esperando uma pela outra para sempre (o famoso **deadlock**). Trancar sempre na mesma ordem elimina esse risco.

Repare também em `NewAccountService`: o serviço **recebe** os repositórios como parâmetros em vez de criá-los. Isso se chama **Injeção de Dependência** e é o que permite ligar "qualquer plugue na tomada".

## Etapa 6 — A camada de INFRAESTRUTURA (banco de dados)

### 6.1 `migrations/001_create_tables.sql` — criando as tabelas

Uma **migration** é um arquivo SQL que cria ou altera a estrutura do banco. Crie `migrations/001_create_tables.sql`:

```sql
-- Este arquivo é executado AUTOMATICAMENTE pelo PostgreSQL na PRIMEIRA vez
-- que o container do banco é criado (graças ao volume em /docker-entrypoint-initdb.d).

-- Tabela de contas
CREATE TABLE IF NOT EXISTS accounts (
    id          UUID        PRIMARY KEY,                         -- identificador único
    owner_name  TEXT        NOT NULL,                            -- nome do titular
    balance     BIGINT      NOT NULL DEFAULT 0 CHECK (balance >= 0), -- saldo em centavos (nunca negativo)
    created_at  TIMESTAMPTZ NOT NULL,                            -- data de criação
    updated_at  TIMESTAMPTZ NOT NULL                             -- data da última alteração
);

-- Tabela de movimentações (extrato)
CREATE TABLE IF NOT EXISTS transactions (
    id          UUID        PRIMARY KEY,
    account_id  UUID        NOT NULL REFERENCES accounts(id),    -- "chave estrangeira": precisa existir em accounts
    type        TEXT        NOT NULL,                            -- DEPOSIT, WITHDRAW, TRANSFER_IN, TRANSFER_OUT
    amount      BIGINT      NOT NULL CHECK (amount > 0),         -- valor em centavos
    created_at  TIMESTAMPTZ NOT NULL
);

-- Índice: deixa rápida a busca do extrato por conta (como o índice de um livro)
CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions (account_id);
```

**O que está acontecendo:**

- `CREATE TABLE` cria uma tabela, e cada linha dentro dela é uma coluna com seu **tipo**: `UUID` (identificador), `TEXT` (texto), `BIGINT` (número inteiro grande), `TIMESTAMPTZ` (data e hora com fuso).
- `NOT NULL` significa "obrigatório".
- `CHECK (balance >= 0)` é uma **rede de segurança**: mesmo que exista um bug no código, o próprio banco se recusa a gravar saldo negativo.
- `REFERENCES accounts(id)` é a chave estrangeira: não dá para registrar uma movimentação para uma conta que não existe.
- `CREATE INDEX` cria um índice, que funciona como o índice remissivo de um livro: deixa a busca do extrato rápida, mesmo com milhões de linhas.

### 6.2 `internal/infrastructure/postgres/db.go` — a conexão

```go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	// O "_" significa: importe este pacote só para ele se registrar.
	// Ele é o "driver", o tradutor entre o Go e o PostgreSQL.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DBTX é "qualquer coisa que consegue executar SQL".
// Tanto a conexão normal (*sql.DB) quanto uma transação (*sql.Tx) servem.
// Graças a isso, os repositórios funcionam dentro e fora de uma UnitOfWork.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Connect abre a conexão com o PostgreSQL.
// Tenta algumas vezes, porque o banco pode demorar uns segundos para ficar pronto.
func Connect(ctx context.Context, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("abrir conexão: %w", err)
	}

	// Limites do "pool" (conjunto de conexões reaproveitáveis).
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	const maxAttempts = 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = db.PingContext(pingCtx)
		cancel()
		if err == nil {
			return db, nil
		}
		log.Printf("banco ainda não respondeu (tentativa %d/%d): %v", attempt, maxAttempts, err)
		time.Sleep(2 * time.Second)
	}

	db.Close()
	return nil, fmt.Errorf("não foi possível conectar ao banco: %w", err)
}
```

**O que está acontecendo:**

- `sql.Open("pgx", ...)` prepara a conexão usando o **driver** `pgx`, o "tradutor" entre Go e PostgreSQL.
- O **pool** de conexões é como uma frota de táxis: em vez de abrir uma conexão nova a cada requisição (lento), reaproveitamos até 10 conexões abertas.
- O laço `for` tenta conectar até 10 vezes, esperando 2 segundos entre as tentativas, porque o PostgreSQL pode levar alguns segundos para ficar pronto ao subir com o Docker.
- A interface `DBTX` é um truque útil: tanto a conexão normal quanto uma transação sabem executar SQL, então os repositórios aceitam qualquer uma das duas.

### 6.3 `internal/infrastructure/postgres/account_repository.go` — o plugue das contas

```go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"banco-api/internal/domain"
)

// AccountRepository é a implementação REAL do contrato domain.AccountRepository,
// usando PostgreSQL. É aqui (e só aqui) que existe SQL relacionado a contas.
type AccountRepository struct {
	db DBTX
}

func NewAccountRepository(db DBTX) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	// $1, $2... são "parâmetros". NUNCA monte SQL juntando textos com dados do
	// usuário: isso abre brecha para um ataque chamado SQL Injection.
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO accounts (id, owner_name, balance, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		account.ID(),
		account.OwnerName(),
		int64(account.Balance()),
		account.CreatedAt(),
		account.UpdatedAt(),
	)
	return err
}

func (r *AccountRepository) FindByID(ctx context.Context, id string) (*domain.Account, error) {
	return r.findOne(ctx,
		`SELECT id, owner_name, balance, created_at, updated_at
		 FROM accounts WHERE id = $1`, id)
}

func (r *AccountRepository) FindByIDForUpdate(ctx context.Context, id string) (*domain.Account, error) {
	// "FOR UPDATE" tranca a linha até o fim da transação. Outra operação que
	// tente mexer nessa mesma conta vai esperar a primeira terminar.
	return r.findOne(ctx,
		`SELECT id, owner_name, balance, created_at, updated_at
		 FROM accounts WHERE id = $1 FOR UPDATE`, id)
}

func (r *AccountRepository) Update(ctx context.Context, account *domain.Account) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE accounts SET balance = $1, updated_at = $2 WHERE id = $3`,
		int64(account.Balance()),
		account.UpdatedAt(),
		account.ID(),
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

// findOne executa a consulta e transforma a linha do banco em uma entidade do domínio.
func (r *AccountRepository) findOne(ctx context.Context, query, id string) (*domain.Account, error) {
	// Se o ID nem tem o formato de UUID, a conta com certeza não existe.
	if _, err := uuid.Parse(id); err != nil {
		return nil, domain.ErrAccountNotFound
	}

	var (
		accountID string
		ownerName string
		balance   int64
		createdAt time.Time
		updatedAt time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&accountID, &ownerName, &balance, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}

	return domain.RestoreAccount(accountID, ownerName, domain.Money(balance), createdAt, updatedAt), nil
}
```

**O que está acontecendo:** este é o **plugue** que encaixa na tomada `domain.AccountRepository`. Em Go, não é preciso escrever "implementa a interface": como a struct tem todos os métodos exigidos (`Create`, `FindByID`, `FindByIDForUpdate`, `Update`), ela encaixa automaticamente.

- `ExecContext` executa comandos que não retornam linhas (`INSERT`, `UPDATE`).
- `QueryRowContext(...).Scan(...)` busca **uma** linha e copia cada coluna para uma variável, na mesma ordem do `SELECT`.
- `sql.ErrNoRows` é o erro que o Go devolve quando a busca não encontra nada. Nós o traduzimos para o erro de negócio `domain.ErrAccountNotFound`. Assim, as camadas de dentro nunca precisam saber que existe SQL.
- **Os `$1`, `$2`...** são parâmetros. Isso protege contra **SQL Injection**, um dos ataques mais comuns da internet, em que alguém envia um texto malicioso tentando executar comandos no seu banco.

### 6.4 `internal/infrastructure/postgres/transaction_repository.go` — o plugue do extrato

```go
package postgres

import (
	"context"

	"banco-api/internal/domain"
)

// TransactionRepository implementa domain.TransactionRepository com PostgreSQL.
type TransactionRepository struct {
	db DBTX
}

func NewTransactionRepository(db DBTX) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, t *domain.Transaction) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO transactions (id, account_id, type, amount, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		t.ID,
		t.AccountID,
		string(t.Type),
		int64(t.Amount),
		t.CreatedAt,
	)
	return err
}

func (r *TransactionRepository) ListByAccountID(ctx context.Context, accountID string) ([]*domain.Transaction, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, account_id, type, amount, created_at
		 FROM transactions
		 WHERE account_id = $1
		 ORDER BY created_at DESC`, accountID)
	if err != nil {
		return nil, err
	}
	// "defer" agenda algo para rodar quando a função terminar.
	// Aqui garantimos que o resultado da consulta seja sempre fechado.
	defer rows.Close()

	transactions := make([]*domain.Transaction, 0)
	for rows.Next() {
		var (
			t      domain.Transaction
			txType string
			amount int64
		)
		if err := rows.Scan(&t.ID, &t.AccountID, &txType, &amount, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Type = domain.TransactionType(txType)
		t.Amount = domain.Money(amount)
		transactions = append(transactions, &t)
	}
	return transactions, rows.Err()
}
```

**O que está acontecendo:** `QueryContext` busca **várias** linhas. O laço `for rows.Next()` passa por cada uma delas e monta a lista. `ORDER BY created_at DESC` ordena da mais recente para a mais antiga, como num extrato de verdade.

### 6.5 `internal/infrastructure/postgres/unit_of_work.go` — o "tudo ou nada" com PostgreSQL

```go
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"banco-api/internal/application"
)

// UnitOfWork implementa application.UnitOfWork usando uma TRANSAÇÃO do PostgreSQL.
//
// Transação de banco de dados = um "pacote" de comandos SQL que é salvo de
// uma vez só (COMMIT) ou descartado inteiro (ROLLBACK).
type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(repos application.Repositories) error) error {
	tx, err := u.db.BeginTx(ctx, nil) // BEGIN: abre o "pacote"
	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}

	// Se algo der errado (ou até um panic), desfaz tudo.
	// Se o Commit já tiver acontecido, este Rollback simplesmente não faz nada.
	defer func() { _ = tx.Rollback() }()

	// Repositórios que usam a transação (tx) em vez da conexão normal.
	repos := application.Repositories{
		Accounts:     NewAccountRepository(tx),
		Transactions: NewTransactionRepository(tx),
	}

	if err := fn(repos); err != nil {
		return err // o defer acima fará o ROLLBACK
	}

	if err := tx.Commit(); err != nil { // COMMIT: salva tudo de uma vez
		return fmt.Errorf("confirmar transação: %w", err)
	}
	return nil
}
```

**O que está acontecendo:** o fluxo é `BEGIN` → executa o trabalho → `COMMIT` (se deu tudo certo) ou `ROLLBACK` (se deu erro). O detalhe mais importante: os repositórios criados aqui recebem `tx` (a transação), e não a conexão normal. Assim, tudo o que eles fizerem faz parte do mesmo "pacote".

## Etapa 7 — A camada de ENTREGA (HTTP com Gin)

Crie os arquivos dentro de `internal/handler/`.

### 7.1 `internal/handler/dto.go` — o formato do JSON

```go
package handler

import (
	"time"

	"banco-api/internal/domain"
)

// DTO = Data Transfer Object ("objeto de transporte de dados").
// São structs que representam o JSON que ENTRA e que SAI da API.
// Elas existem para que o formato do JSON seja independente das entidades do domínio.
//
// As "tags" entre crases dizem:
//   json:"owner_name"     -> nome do campo no JSON
//   binding:"required"    -> o Gin recusa a requisição se o campo não vier

// ----- Entrada (request) -----

type CreateAccountRequest struct {
	OwnerName string `json:"owner_name" binding:"required"`
}

// AmountRequest é usado em depósito e saque. O valor é em CENTAVOS.
type AmountRequest struct {
	Amount int64 `json:"amount"`
}

type TransferRequest struct {
	FromAccountID string `json:"from_account_id" binding:"required"`
	ToAccountID   string `json:"to_account_id" binding:"required"`
	Amount        int64  `json:"amount"`
}

// ----- Saída (response) -----

type AccountResponse struct {
	ID               string    `json:"id"`
	OwnerName        string    `json:"owner_name"`
	Balance          int64     `json:"balance"`
	BalanceFormatted string    `json:"balance_formatted"`
	CreatedAt        time.Time `json:"created_at"`
}

type TransactionResponse struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	Amount          int64     `json:"amount"`
	AmountFormatted string    `json:"amount_formatted"`
	CreatedAt       time.Time `json:"created_at"`
}

// Funções que convertem entidades do domínio em DTOs de saída.

func toAccountResponse(a *domain.Account) AccountResponse {
	return AccountResponse{
		ID:               a.ID(),
		OwnerName:        a.OwnerName(),
		Balance:          int64(a.Balance()),
		BalanceFormatted: a.Balance().String(),
		CreatedAt:        a.CreatedAt(),
	}
}

func toTransactionResponses(list []*domain.Transaction) []TransactionResponse {
	result := make([]TransactionResponse, 0, len(list))
	for _, t := range list {
		result = append(result, TransactionResponse{
			ID:              t.ID,
			Type:            string(t.Type),
			Amount:          int64(t.Amount),
			AmountFormatted: t.Amount.String(),
			CreatedAt:       t.CreatedAt,
		})
	}
	return result
}
```

**O que está acontecendo:** DTOs definem **como o JSON se parece** do lado de fora. Por que não devolver a entidade `Account` direto? Porque o formato público da API não deve ficar preso ao formato interno do domínio. Se um dia você adicionar um campo interno na conta, ele não vaza sem querer para o JSON. Aproveitamos também para devolver o valor formatado (`"R$ 59,50"`), útil para quem consome a API.

Repare que `amount` **não** tem `binding:"required"`. Foi de propósito: a regra "o valor deve ser maior que zero" é **de negócio**, então quem responde por ela é o domínio. O handler só verifica se o formato está correto (por exemplo, se veio um número e não um texto).

### 7.2 `internal/handler/errors.go` — traduzindo erros em códigos HTTP

```go
package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"banco-api/internal/domain"
)

// respondError traduz erros de NEGÓCIO em respostas HTTP.
// O domínio diz "saldo insuficiente"; é aqui que decidimos que isso vira o status 422.
func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}) // 404

	case errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrInvalidOwnerName),
		errors.Is(err, domain.ErrSameAccountTransfer):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // 400

	case errors.Is(err, domain.ErrInsufficientFunds):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()}) // 422

	default:
		// Erro inesperado (ex.: banco fora do ar). Registramos o detalhe no log,
		// mas não mostramos ao cliente, pois pode conter informação sensível.
		log.Printf("erro inesperado: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno do servidor"}) // 500
	}
}
```

**O que está acontecendo:** o domínio diz "saldo insuficiente"; é **aqui** que decidimos que isso vira o status `422`. Se amanhã você criar uma versão da API com outra tecnologia (gRPC, por exemplo), só esta tradução muda; o domínio continua igual. `gin.H` é só um atalho do Gin para montar um JSON rapidamente: `gin.H{"error": "..."}` vira `{"error": "..."}`.

### 7.3 `internal/handler/account_handler.go` — os "garçons"

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banco-api/internal/application"
	"banco-api/internal/domain"
)

// AccountHandler recebe as requisições HTTP e chama os casos de uso.
// Cada método segue sempre o mesmo roteiro:
//  1. ler e validar o formato da entrada (JSON, parâmetros da URL);
//  2. chamar o caso de uso;
//  3. transformar o resultado (ou o erro) em uma resposta HTTP.
type AccountHandler struct {
	service *application.AccountService
}

func NewAccountHandler(service *application.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

// POST /accounts
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida: " + err.Error()})
		return
	}

	account, err := h.service.CreateAccount(c.Request.Context(), req.OwnerName)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toAccountResponse(account)) // 201 = criado
}

// GET /accounts/:id
func (h *AccountHandler) GetAccount(c *gin.Context) {
	id := c.Param("id") // pega o ":id" da URL

	account, err := h.service.GetAccount(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toAccountResponse(account)) // 200 = ok
}

// POST /accounts/:id/deposit
func (h *AccountHandler) Deposit(c *gin.Context) {
	var req AmountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida: " + err.Error()})
		return
	}

	account, err := h.service.Deposit(c.Request.Context(), c.Param("id"), domain.Money(req.Amount))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toAccountResponse(account))
}

// POST /accounts/:id/withdraw
func (h *AccountHandler) Withdraw(c *gin.Context) {
	var req AmountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida: " + err.Error()})
		return
	}

	account, err := h.service.Withdraw(c.Request.Context(), c.Param("id"), domain.Money(req.Amount))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toAccountResponse(account))
}

// POST /transfers
func (h *AccountHandler) Transfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida: " + err.Error()})
		return
	}

	err := h.service.Transfer(c.Request.Context(), req.FromAccountID, req.ToAccountID, domain.Money(req.Amount))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transferência realizada com sucesso"})
}

// GET /accounts/:id/transactions
func (h *AccountHandler) ListTransactions(c *gin.Context) {
	transactions, err := h.service.ListTransactions(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTransactionResponses(transactions))
}
```

**O que está acontecendo:** cada função recebe um `*gin.Context`, que carrega **tudo sobre a requisição** (a URL, o corpo, os cabeçalhos) e é por onde enviamos a resposta.

- `c.ShouldBindJSON(&req)` lê o JSON do corpo e preenche a struct. Se o JSON estiver quebrado, devolvemos `400`.
- `c.Param("id")` pega o pedaço `:id` da rota.
- `c.Request.Context()` é o contexto da requisição, que repassamos ao caso de uso.
- `c.JSON(status, dados)` envia a resposta.

Repare como os handlers são "burros" de propósito: não há nenhuma regra de negócio aqui, só tradução entre HTTP e os casos de uso.

### 7.4 `internal/handler/router.go` — a tabela de rotas

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter cria o "roteador": a tabela que liga cada URL + método HTTP
// à função que vai tratá-la.
func NewRouter(h *AccountHandler) *gin.Engine {
	// gin.Default() já vem com Logger (mostra cada requisição no terminal)
	// e Recovery (evita que a API caia se acontecer um panic).
	router := gin.Default()

	// Rota simples para verificar se a API está de pé.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Group agrupa rotas com o mesmo prefixo "/accounts".
	accounts := router.Group("/accounts")
	{
		accounts.POST("", h.CreateAccount)                    // POST /accounts
		accounts.GET("/:id", h.GetAccount)                    // GET  /accounts/{id}
		accounts.POST("/:id/deposit", h.Deposit)              // POST /accounts/{id}/deposit
		accounts.POST("/:id/withdraw", h.Withdraw)            // POST /accounts/{id}/withdraw
		accounts.GET("/:id/transactions", h.ListTransactions) // GET  /accounts/{id}/transactions
	}

	router.POST("/transfers", h.Transfer) // POST /transfers

	return router
}
```

**O que está acontecendo:** aqui ligamos cada **método HTTP + rota** a uma função. `:id` indica uma parte variável da URL. `router.Group("/accounts")` evita repetir o prefixo `/accounts` em cada rota.

## Etapa 8 — O `main.go`: montando as peças

Crie `cmd/api/main.go`:

```go
package main

import (
	"context"
	"log"
	"os"

	"banco-api/internal/application"
	"banco-api/internal/handler"
	"banco-api/internal/infrastructure/postgres"
)

// main é o ponto de partida do programa.
// Ela é a "sala de montagem": cria cada peça e encaixa uma na outra.
// É o ÚNICO lugar que conhece todas as camadas ao mesmo tempo.
func main() {
	// 1. Lê as configurações das variáveis de ambiente (definidas no docker-compose.yml).
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("a variável de ambiente DATABASE_URL é obrigatória")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. Conecta ao banco de dados.
	db, err := postgres.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 3. Infraestrutura: implementações concretas dos repositórios.
	accountRepo := postgres.NewAccountRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	uow := postgres.NewUnitOfWork(db)

	// 4. Aplicação: casos de uso, recebendo os repositórios.
	accountService := application.NewAccountService(accountRepo, transactionRepo, uow)

	// 5. Entrega (HTTP): handlers e rotas, recebendo os casos de uso.
	accountHandler := handler.NewAccountHandler(accountService)
	router := handler.NewRouter(accountHandler)

	// 6. Liga o servidor.
	log.Printf("API rodando na porta %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
```

**O que está acontecendo:** o `main.go` é a **sala de montagem**. É o único lugar que conhece todas as camadas, e ele encaixa uma na outra de fora para dentro: cria os repositórios (plugues), entrega-os ao serviço (que só conhece as tomadas), entrega o serviço ao handler e o handler ao roteador. Por fim, liga o servidor.

`os.Getenv("DATABASE_URL")` lê uma **variável de ambiente**, uma configuração passada "de fora" para o programa (quem vai defini-la é o `docker-compose.yml`). Assim, o endereço e a senha do banco não ficam gravados no código.

## Etapa 9 — Baixar as bibliotecas (dependências)

O código usa três bibliotecas externas: **Gin**, **pgx** (driver do PostgreSQL) e **uuid**. O comando `go mod tidy` lê todos os `import` do projeto, baixa o que falta e atualiza o `go.mod`, além de criar o `go.sum` (uma espécie de "lacre de segurança" que garante que as bibliotecas não foram adulteradas).

**Mac, Linux ou Git Bash:**

```bash
docker run --rm -v "$(pwd)":/app -w /app golang:1.27-alpine go mod tidy
```

**Windows PowerShell:**

```powershell
docker run --rm -v "${PWD}:/app" -w /app golang:1.27-alpine go mod tidy
```

Ao final, confira: o `go.mod` agora deve listar as bibliotecas em um bloco `require (...)`, e deve existir um arquivo `go.sum`.

> **Linux:** se os arquivos criados pelo container ficarem com dono `root`, rode `sudo chown -R $USER .` na pasta do projeto.

## Etapa 10 — O `Dockerfile`: a receita da imagem da API

Crie um arquivo chamado exatamente `Dockerfile` (sem extensão) na raiz do projeto:

```dockerfile
# ============================================================
# ETAPA 1 - "builder": uma imagem com o Go instalado, só para COMPILAR
# ============================================================
FROM golang:1.27-alpine AS builder

# Pasta de trabalho dentro do container
WORKDIR /app

# Copia primeiro só a lista de dependências e baixa tudo.
# Assim o Docker reaproveita esse passo (cache) enquanto as dependências não mudarem.
COPY go.mod go.sum ./
RUN go mod download

# Agora copia o resto do código
COPY . .

# Compila o programa, gerando um único arquivo executável chamado "api"
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

# ============================================================
# ETAPA 2 - imagem final: bem pequena, só com o executável
# ============================================================
FROM alpine:3.22

WORKDIR /app

# Pega APENAS o executável da etapa anterior (o Go não vai junto)
COPY --from=builder /api /app/api

# Documenta que a API usa a porta 8080
EXPOSE 8080

# Comando executado quando o container inicia
CMD ["/app/api"]
```

**O que está acontecendo:** é um **build em duas etapas** (*multi-stage build*):

1. **Etapa "builder":** usamos uma imagem grande, com o Go instalado, só para **compilar** o código num único arquivo executável. O truque de copiar `go.mod`/`go.sum` antes do resto do código faz o Docker guardar em cache as bibliotecas baixadas: enquanto você não mudar as dependências, os próximos builds ficam bem mais rápidos.
2. **Etapa final:** partimos de uma imagem Linux minúscula (Alpine, com poucos MB) e copiamos **só o executável**. O resultado é uma imagem pequena, rápida de baixar e mais segura, porque não carrega ferramentas desnecessárias.

`CGO_ENABLED=0` gera um executável que não depende de nada do sistema, então ele roda em qualquer Linux.

## Etapa 11 — O `docker-compose.yml`: orquestrando tudo

Crie `docker-compose.yml` na raiz do projeto:

```yaml
# O docker-compose descreve TODOS os serviços do projeto e como eles se conectam.
# Com um único comando (docker compose up) tudo sobe junto.

services:

  # ---------- Banco de dados PostgreSQL ----------
  db:
    image: postgres:17-alpine          # imagem oficial pronta do PostgreSQL
    container_name: banco-db
    environment:                       # configurações iniciais do banco
      POSTGRES_USER: banco
      POSTGRES_PASSWORD: banco123
      POSTGRES_DB: banco_db
    ports:
      - "5432:5432"                    # porta_do_seu_pc:porta_do_container
    volumes:
      - pgdata:/var/lib/postgresql/data             # guarda os dados mesmo se o container for removido
      - ./migrations:/docker-entrypoint-initdb.d    # scripts SQL executados na 1ª inicialização
    healthcheck:                       # como o Docker descobre se o banco está pronto
      test: ["CMD-SHELL", "pg_isready -U banco -d banco_db"]
      interval: 5s
      timeout: 5s
      retries: 10

  # ---------- Nossa API em Go ----------
  api:
    build: .                           # constrói a imagem usando o Dockerfile desta pasta
    container_name: banco-api
    environment:
      # Repare no "@db:5432": dentro do Docker, o nome do serviço ("db") funciona como endereço.
      DATABASE_URL: postgres://banco:banco123@db:5432/banco_db?sslmode=disable
      PORT: "8080"
    ports:
      - "8080:8080"
    depends_on:
      db:
        condition: service_healthy     # só inicia a API depois que o banco estiver saudável
    restart: on-failure                # se a API cair por erro, o Docker tenta subir de novo

  # ---------- Adminer: tela web para ver o banco de dados ----------
  adminer:
    image: adminer
    container_name: banco-adminer
    ports:
      - "8081:8080"
    depends_on:
      - db

# Volumes nomeados: "HDs virtuais" gerenciados pelo Docker
volumes:
  pgdata:
```

**O que está acontecendo, serviço por serviço:**

**`db` (PostgreSQL):** usa a imagem oficial pronta. As variáveis `POSTGRES_USER`, `POSTGRES_PASSWORD` e `POSTGRES_DB` criam o usuário, a senha e o banco automaticamente na primeira vez. Os volumes fazem duas coisas: `pgdata` guarda os dados de forma permanente, e `./migrations` coloca nosso SQL numa pasta especial (`/docker-entrypoint-initdb.d`) cujos scripts o PostgreSQL executa sozinho **na primeira inicialização**. O `healthcheck` diz ao Docker como perguntar "banco, você está pronto?".

**`api` (nossa aplicação):** `build: .` manda o Compose construir a imagem com o nosso `Dockerfile`. `depends_on` com `service_healthy` faz a API esperar o banco estar saudável antes de iniciar.

Um detalhe importante é o endereço do banco: `postgres://banco:banco123@db:5432/banco_db`. Ele se lê como `postgres://USUÁRIO:SENHA@ENDEREÇO:PORTA/NOME_DO_BANCO`. O endereço é `db`, o **nome do serviço**, e não `localhost`! O Compose cria uma rede interna em que cada serviço encontra o outro pelo nome. Para a API, `localhost` seria o próprio container da API.

**`ports: "8080:8080"`** se lê como `porta_no_seu_computador:porta_dentro_do_container`. É isso que permite acessar `http://localhost:8080` do seu navegador.

**`adminer`:** uma interface web leve para você **ver o banco de dados** sem instalar nada.

**`volumes: pgdata:`** no final declara o volume nomeado.

> ⚠️ **Senhas no arquivo:** colocar senhas direto no `docker-compose.yml` é aceitável para **estudo**. Em projetos reais, use um arquivo `.env` (que não vai para o Git) ou um gerenciador de segredos.

Crie também o `.dockerignore` na raiz (ele diz o que **não** deve ser copiado para dentro da imagem):

```text
# Arquivos que NÃO precisam ir para dentro da imagem Docker
.git
*.md
docker-compose.yml
```

## Etapa 12 — Subir tudo! 🚀

Com o Docker Desktop aberto, na pasta do projeto, rode:

```bash
docker compose up --build
```

- `up` = sobe todos os serviços do `docker-compose.yml`;
- `--build` = (re)constrói a imagem da API antes de subir. Use sempre que mudar o código Go.

Na primeira vez, vai demorar alguns minutos (o Docker baixa as imagens e compila o projeto). Você vai ver os logs dos três serviços misturados. Espere aparecer algo como:

```
banco-api  | [GIN-debug] POST   /accounts   --> banco-api/internal/handler.(*AccountHandler).CreateAccount-fm (3 handlers)
banco-api  | [GIN-debug] GET    /accounts/:id --> ...
banco-api  | 2026/09/10 19:42:48 API rodando na porta 8080
```

O Gin mostra a tabela de rotas e um aviso `[WARNING] Running in "debug" mode`. Isso é normal em desenvolvimento. Em produção, você definiria a variável `GIN_MODE=release`.

Esse terminal ficará "preso" mostrando os logs. Para testar, abra **outro terminal**. Para parar tudo, volte a este e aperte `Ctrl + C`.

> Prefere o terminal livre? Use `docker compose up --build -d` (o `-d` roda em segundo plano) e veja os logs com `docker compose logs -f api`.

## Etapa 13 — Testar a API

### Opção A: pelo VS Code (mais fácil, funciona em qualquer sistema)

Crie o arquivo `requests.http` na raiz do projeto com o conteúdo abaixo. Com a extensão **REST Client** instalada, aparece um link **"Send Request"** em cima de cada requisição; é só clicar.

```http
# Arquivo para testar a API pelo VS Code.
# Instale a extensão "REST Client" (autor: Huachao Mao) e clique em "Send Request"
# que aparece em cima de cada requisição.

@baseUrl = http://localhost:8080

# Depois de criar as contas, cole os IDs retornados aqui:
@contaA = COLE-AQUI-O-ID-DA-CONTA-A
@contaB = COLE-AQUI-O-ID-DA-CONTA-B

### 1. Verificar se a API está no ar
GET {{baseUrl}}/health

### 2. Criar a conta A
POST {{baseUrl}}/accounts
Content-Type: application/json

{
  "owner_name": "Maria Silva"
}

### 3. Criar a conta B
POST {{baseUrl}}/accounts
Content-Type: application/json

{
  "owner_name": "João Souza"
}

### 4. Consultar a conta A
GET {{baseUrl}}/accounts/{{contaA}}

### 5. Depositar R$ 100,00 na conta A (valor em centavos)
POST {{baseUrl}}/accounts/{{contaA}}/deposit
Content-Type: application/json

{
  "amount": 10000
}

### 6. Sacar R$ 25,50 da conta A
POST {{baseUrl}}/accounts/{{contaA}}/withdraw
Content-Type: application/json

{
  "amount": 2550
}

### 7. Transferir R$ 15,00 da conta A para a conta B
POST {{baseUrl}}/transfers
Content-Type: application/json

{
  "from_account_id": "{{contaA}}",
  "to_account_id": "{{contaB}}",
  "amount": 1500
}

### 8. Ver o extrato da conta A
GET {{baseUrl}}/accounts/{{contaA}}/transactions

### 9. Testar um erro: sacar mais do que o saldo
POST {{baseUrl}}/accounts/{{contaA}}/withdraw
Content-Type: application/json

{
  "amount": 99999999
}
```

Fluxo: rode a requisição 2 e a 3, copie os `id` que voltarem para as linhas `@contaA` e `@contaB` no topo do arquivo, e siga com as outras.

### Opção B: pelo terminal com `curl` (Mac, Linux ou Git Bash)

```bash
# 1. A API está no ar?
curl http://localhost:8080/health
# {"status":"ok"}

# 2. Criar uma conta
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"owner_name": "Maria Silva"}'
# {"id":"9a67b859-...","owner_name":"Maria Silva","balance":0,"balance_formatted":"R$ 0,00",...}
```

Copie o `id` retornado e guarde numa variável do terminal para facilitar (troque pelo seu ID):

```bash
CONTA_A=9a67b859-8c59-4561-84c8-da8a6eec82ad
```

Crie uma segunda conta do mesmo jeito (com outro nome) e guarde em `CONTA_B`. Então:

```bash
# Depositar R$ 100,00 (10000 centavos)
curl -X POST http://localhost:8080/accounts/$CONTA_A/deposit \
  -H "Content-Type: application/json" -d '{"amount": 10000}'
# {... "balance":10000,"balance_formatted":"R$ 100,00" ...}

# Sacar R$ 25,50
curl -X POST http://localhost:8080/accounts/$CONTA_A/withdraw \
  -H "Content-Type: application/json" -d '{"amount": 2550}'
# {... "balance":7450,"balance_formatted":"R$ 74,50" ...}

# Transferir R$ 15,00 de A para B
curl -X POST http://localhost:8080/transfers \
  -H "Content-Type: application/json" \
  -d "{\"from_account_id\": \"$CONTA_A\", \"to_account_id\": \"$CONTA_B\", \"amount\": 1500}"
# {"message":"transferência realizada com sucesso"}

# Ver o extrato de A
curl http://localhost:8080/accounts/$CONTA_A/transactions
# [{"type":"TRANSFER_OUT","amount":1500,...},{"type":"WITHDRAW",...},{"type":"DEPOSIT",...}]
```

### Testando as regras de negócio (os erros)

| Teste | Resposta esperada |
|---|---|
| Sacar mais do que o saldo | `422` `{"error":"saldo insuficiente"}` |
| Depositar `0` ou valor negativo | `400` `{"error":"o valor deve ser maior que zero"}` |
| Transferir para a mesma conta | `400` `{"error":"não é possível transferir para a mesma conta"}` |
| Consultar uma conta que não existe | `404` `{"error":"conta não encontrada"}` |
| Criar conta com nome só de espaços | `400` `{"error":"o nome do titular é obrigatório"}` |
| Enviar `"amount": "abc"` | `400` `{"error":"requisição inválida: ..."}` |

Todos esses comportamentos foram verificados, incluindo um teste com 20 saques simultâneos na mesma conta: só os que cabiam no saldo foram aceitos, e o saldo nunca ficou errado, graças ao `FOR UPDATE` e à `UnitOfWork`.

## Etapa 14 — Ver o banco de dados com o Adminer

Abra http://localhost:8081 no navegador e preencha:

| Campo | Valor |
|---|---|
| Sistema | PostgreSQL |
| Servidor | `db` |
| Usuário | `banco` |
| Senha | `banco123` |
| Base de dados | `banco_db` |

Clique nas tabelas `accounts` e `transactions` e veja os dados que você criou pela API. Experimente o botão **"Comando SQL"** e rode, por exemplo:

```sql
SELECT owner_name, balance FROM accounts ORDER BY balance DESC;
```

## Etapa 15 — Rodar os testes automáticos do domínio

```bash
# Mac, Linux ou Git Bash
docker run --rm -v "$(pwd)":/app -w /app golang:1.27-alpine go test ./...

# Windows PowerShell
docker run --rm -v "${PWD}:/app" -w /app golang:1.27-alpine go test ./...
```

Resultado esperado:

```
ok      banco-api/internal/domain       0.003s
```

(As outras pastas aparecem como `[no test files]`, porque só escrevemos testes para o domínio.)

Pare um instante para perceber o que acabou de acontecer: você testou todas as regras do banco **sem banco de dados e sem subir a API**. É exatamente para isso que servem o DDD e a Clean Architecture.

---

# PARTE 3 — Comandos do Docker que você vai usar no dia a dia

| Comando | O que faz |
|---|---|
| `docker compose up --build` | Constrói e sobe tudo, mostrando os logs |
| `docker compose up --build -d` | Mesma coisa, em segundo plano |
| `docker compose ps` | Lista os containers e se estão saudáveis |
| `docker compose logs -f api` | Acompanha os logs só da API (`Ctrl + C` para sair) |
| `docker compose restart api` | Reinicia só a API |
| `docker compose stop` | Pausa os containers (os dados ficam) |
| `docker compose down` | Remove os containers (os dados **ficam** no volume) |
| `docker compose down -v` | Remove os containers **e apaga o banco** (o volume). Na próxima subida, as migrations rodam de novo |
| `docker compose exec db psql -U banco -d banco_db` | Abre um terminal SQL dentro do container do banco (saia com `\q`) |

**Fluxo típico de desenvolvimento:** alterou código Go → `docker compose up --build`. Alterou o SQL da migration → `docker compose down -v` e depois `docker compose up --build` (atenção: isso apaga os dados).

---

# PARTE 4 — Problemas comuns e soluções

**"Cannot connect to the Docker daemon" / "error during connect"**
O Docker Desktop não está aberto. Abra-o e espere o ícone indicar que está rodando.

**"port is already allocated" ou "address already in use"**
Alguma coisa no seu computador já usa essa porta (é comum já existir um PostgreSQL instalado usando a 5432). No `docker-compose.yml`, troque **o número da esquerda**: por exemplo, `"5433:5432"` ou `"8090:8080"`. O da direita é a porta dentro do container e não muda.

**"COPY failed: ... go.sum: not found"**
Você pulou a Etapa 9. Rode o `go mod tidy` e tente de novo.

**"go.mod requires go >= 1.XX (running go 1.YY; GOTOOLCHAIN=local)"**
Alguma biblioteca exige uma versão do Go mais nova que a da imagem. Troque `golang:1.27-alpine` no `Dockerfile` por uma versão igual ou maior que a pedida na mensagem.

**Mudei o código e nada mudou**
Você esqueceu o `--build`. Rode `docker compose up --build`.

**Mudei o arquivo SQL e as tabelas não mudaram**
Os scripts de `/docker-entrypoint-initdb.d` só rodam quando o banco é criado **pela primeira vez**. Rode `docker compose down -v` (apaga os dados) e suba de novo.

**A API mostra "banco ainda não respondeu (tentativa X/10)"**
É normal aparecer uma ou duas vezes enquanto o PostgreSQL inicia. Se chegar a 10, confira os logs do banco com `docker compose logs db`.

**No Windows, o `curl` dá erros estranhos**
No PowerShell, `curl` é um apelido de outro comando e as aspas do JSON se comportam de outro jeito. Use o `requests.http` (Opção A) ou o Git Bash.

---

# PARTE 5 — Próximos passos (e o que falta para produção)

Este projeto é **didático**. Ele já tem boas práticas importantes (dinheiro em centavos, transações, trava contra concorrência, separação em camadas), mas antes de algo parecido ir para produção faltariam, entre outros:

1. **Autenticação e autorização (o mais importante!):** hoje, qualquer pessoa que souber o ID de uma conta consegue sacar dela. O próximo passo natural é adicionar login com **JWT** e um *middleware* do Gin que verifique quem está chamando.
2. **Ferramenta de migrations:** trocar o `/docker-entrypoint-initdb.d` por uma ferramenta como o **golang-migrate** ou o **goose**, que aplica mudanças de forma incremental sem apagar o banco.
3. **Testes dos casos de uso com repositórios falsos (mocks)** e **testes de integração** com um banco real (a biblioteca **testcontainers-go** sobe um PostgreSQL só para os testes).
4. **Idempotência:** se a internet cair logo depois do cliente pedir uma transferência, ele pode tentar de novo e transferir duas vezes. Bancos reais usam uma *chave de idempotência* enviada pelo cliente.
5. **Graceful shutdown:** ao desligar, a API termina as requisições em andamento antes de sair.
6. **Documentação automática** da API com **Swagger/OpenAPI** (biblioteca `swaggo`).
7. **Logs estruturados** (pacote `log/slog`) e configuração com arquivo `.env`.
8. **Mais DDD:** um Objeto de Valor `CPF` com validação, e **eventos de domínio** (por exemplo, "TransferênciaRealizada", para enviar uma notificação).

Uma sugestão de estudo: escolha **um** desses itens por vez, implemente e sempre se pergunte "em qual camada isso mora?". Se a resposta for "é regra do negócio", vai para o domínio. Se for "é detalhe técnico", vai para fora. Essa pergunta é a essência de tudo o que você aprendeu aqui.
