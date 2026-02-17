# DockScope

![License](https://img.shields.io/github/license/dockscope/dockscope)
![Go Version](https://img.shields.io/github/go-mod/go-version/dockscope/dockscope)
![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)

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

### Um comando: ambiente fullstack (recomendado)

Para subir **backend (Go) e frontend (React)** em desenvolvimento com um único comando:

```bash
git clone https://github.com/dockscope/dockscope.git
cd dockscope
go mod tidy
cd web && npm install && cd ..
./run.sh
```

O **`run.sh`** é o script de automação do projeto: gere o ciclo de vida dos dois serviços e garante que a API esteja disponível antes de lançar o Vite. Assim, ao abrir `http://localhost:5173`, o dashboard já consegue falar com o backend em `http://localhost:8080` sem erros de conexão.

**Nota técnica:** O script trata o encerramento gracioso (SIGINT/Ctrl+C) de ambos os processos. Ao sair, o trap encerra o backend e evita processos “zumbis” nas portas 8080 ou 5173, para que possas voltar a executar `./run.sh` sem conflito de portas.

Esta abordagem foi pensada para **desenvolvimento nativo e depuração em tempo real**: um único terminal, dependência da API resolvida automaticamente e desligamento limpo.

### Backend (API) só

```bash
go run ./cmd/dockscope
```

A API fica em `http://localhost:8080`. Opções: `--addr=:9090`, `--cli` (listagem no terminal), `--cli --all`, `-v` (debug).

### Frontend (dashboard) só

Com o backend já em execução noutro terminal:

```bash
cd web && npm run dev
```

Abrir `http://localhost:5173`. O proxy do Vite envia `/api` para `:8080`; ajuste `web/vite.config.ts` se a API estiver noutra porta.

### Build para produção

```bash
go build -o dockscope ./cmd/dockscope
cd web && npm run build
```

Os estáticos ficam em `web/dist`; sirva-os com um servidor HTTP e aponte as chamadas à API para o backend.

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
