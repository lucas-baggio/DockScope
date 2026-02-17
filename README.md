# DockScope

**DockScope** é um dashboard open source de observabilidade e gestão para Docker: métricas em tempo real, logs, ações de controle e visão consolidada do host, com backend em Go e interface em React.

## Funcionalidades

- **Dashboard** — Visão geral: total de containers (ativos/parados), CPU e memória agregados, imagens e volumes; gráfico de distribuição de memória; tabela de containers com filtro e ações rápidas (iniciar/parar).
- **Métricas em tempo real** — CPU e memória por container via WebSocket, com gráficos no modal de detalhes.
- **Logs em stream** — Stdout/stderr de cada container em tempo real.
- **Ações de controle** — Start, stop, restart, pause e unpause a partir da interface.
- **CLI** — Listagem de containers, imagens e volumes no terminal (`--cli`).

## Requisitos

- **Go** 1.21+
- **Node.js** 18+ (para o frontend)
- **Docker** — Daemon acessível (socket padrão ou variável `DOCKER_HOST`)

## Instalação e execução

### Subir backend e frontend (recomendado)

Na raiz do repositório, com dependências já instaladas (`go mod tidy` e `cd web && npm install`):

```bash
./run.sh
```

O script inicia o backend (API em `:8080`), espera ficar pronto e depois inicia o frontend (Vite em `:5173`). Ao sair (Ctrl+C), encerra os dois.

### Backend (API) só

```bash
git clone https://github.com/dockscope/dockscope.git
cd dockscope
go mod tidy
go run ./cmd/dockscope
```

A API fica disponível em `http://localhost:8080`.

Opções úteis:

- `--addr=:9090` — Alterar endereço da API
- `--cli` — Modo terminal: listar containers, imagens e volumes e sair
- `--cli --all` — Incluir containers parados na listagem CLI
- `-v` — Logs em modo debug

### Frontend (dashboard)

Com o backend em execução:

```bash
cd web
npm install
npm run dev
```

Abrir `http://localhost:5173`. O Vite faz proxy de `/api` para `http://localhost:8080`; ajuste a porta em `web/vite.config.ts` se a API estiver noutro endereço.

### Build para produção

```bash
go build -o dockscope ./cmd/dockscope
cd web && npm run build
```

O frontend gera os ficheiros estáticos em `web/dist`; sirva-os com qualquer servidor HTTP e garanta que as chamadas à API apontam para o backend.

## API

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/api/health` | Health check |
| GET | `/api/containers` | Lista containers (`?all=true` inclui parados) |
| POST | `/api/containers/{id}/action` | Ação: `{"action":"start"\|"stop"\|"restart"\|"pause"\|"unpause"}` |
| GET | `/api/system/summary` | Sumário do sistema (contagens, CPU/RAM, top por memória) |
| GET | `/api/images` | Lista imagens |
| GET | `/api/volumes` | Lista volumes |
| GET | `/api/stats/{id}` | WebSocket — métricas (CPU, RAM) em tempo real |
| GET | `/api/logs/{id}` | WebSocket — logs (stdout/stderr) em tempo real |

Respostas em JSON. CORS permitido para desenvolvimento.

## Estrutura do projeto

```
cmd/dockscope/           # Entrada da aplicação (CLI e servidor HTTP)
internal/
  domain/                # Entidades e interfaces (sem dependências externas)
  usecase/               # Casos de uso
  infrastructure/
    docker/              # Implementação com Docker SDK
    api/                 # Servidor HTTP e WebSockets
web/                     # Frontend React (Vite, Tailwind, Recharts)
```

Arquitetura em camadas: domínio independente de Docker e HTTP; casos de uso orquestram a lógica; infraestrutura implementa repositórios e API.

## Desenvolvimento

### Testes (Go)

```bash
go test ./...
```

### Lint e build do frontend

```bash
cd web && npm run lint && npm run build
```

### Conexão ao Docker

Por defeito é usado o socket `unix:///var/run/docker.sock`. Para outro host:

```bash
export DOCKER_HOST=tcp://host:2375
```

## Contribuir

Contribuições são bem-vindas: issues para bugs e ideias, pull requests para correções e melhorias. Mantenha o código alinhado ao estilo existente e garanta que os testes passem.

## Licença

MIT License — ver [LICENSE](LICENSE). Copyright (c) 2026 Baggio.
