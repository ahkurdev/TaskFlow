# TaskFlow

A full-stack Kanban board app (Trello-like: Boards -> Lists -> Cards) built to
demonstrate a production-style Go backend paired with a React frontend,
including JWT auth, ownership-scoped multi-tenant data, drag-and-drop, and
real-time sync over WebSocket.

## Quick start

You need 3 terminals running at the same time (in this order):

| # | Command | Folder | Notes |
|---|---|---|---|
| 1 | `docker compose up -d` | `Golang/Web/TaskFlow` | Starts Postgres. Only needs to be run once (leave the container running); it survives closing the terminal. |
| 2 | `go run ./cmd/api` | `Golang/Web/TaskFlow/backend` | Starts the API on `:8081`. Keep this terminal open. |
| 3 | `npm run dev` | `Golang/Web/TaskFlow/frontend` | Starts the web app on `:5173`. Keep this terminal open. |

Then open **http://localhost:5173** in your browser.

First time only, before step 2/3 above:

```sh
# from Golang/Web/TaskFlow/backend
sh scripts/migrate.sh   # creates the database tables
cp .env.example .env    # copy default config
```

```sh
# from Golang/Web/TaskFlow/frontend
npm install              # install dependencies
```

See [Running locally](#running-locally) below for the full explanation of
each step.

## Stack

- **Backend**: Go, [chi](https://github.com/go-chi/chi) router, PostgreSQL
  ([pgx](https://github.com/jackc/pgx)), JWT auth, WebSocket
  ([gorilla/websocket](https://github.com/gorilla/websocket))
- **Frontend**: React + TypeScript (Vite), Tailwind CSS, drag-and-drop via
  [dnd-kit](https://dndkit.com/)
- **Database**: PostgreSQL 16 (via Docker Compose)

## Architecture

The backend follows strict layering:

```
handler -> service -> repository -> database
```

- **handler**: parses HTTP requests, calls a service, writes the response.
  No business logic, no direct DB access.
- **service**: business logic, ownership checks (a board/list/card must
  belong to the requesting user), transactional reordering, and broadcasting
  real-time events.
- **repository**: raw SQL against PostgreSQL via `pgx`. Repositories accept
  a `Querier` interface satisfied by both the connection pool and a
  transaction, so the same repository code works inside or outside a `tx`.

```
backend/
├── cmd/api/main.go        entrypoint: config, DB, routes
├── internal/
│   ├── config/            env loading
│   ├── database/          pgx pool setup
│   ├── handler/           HTTP handlers (health, auth, board, list, card, ws)
│   ├── middleware/         JWT auth middleware
│   ├── model/              domain structs
│   ├── repository/         SQL access (user, board, list, card)
│   ├── service/            business logic + ownership + reordering + tests
│   └── ws/                 WebSocket hub (pub/sub per board)
├── migrations/             plain SQL, numbered
└── scripts/migrate.sh      applies migrations to the running container

frontend/
└── src/
    ├── api/                 typed fetch client (auth, boards, lists, cards)
    ├── context/AuthContext.tsx
    ├── hooks/useBoardSocket.ts   WebSocket subscription -> refetch
    ├── components/          KanbanList, KanbanCard, ProtectedRoute
    └── pages/               Login, Register, Boards, BoardDetail
```

## Features

- Email/password auth with bcrypt-hashed passwords and JWT sessions
- Boards, Lists, and Cards with full CRUD
- Drag-and-drop: reorder lists, reorder cards within a list, and move cards
  across lists (position resequencing handled transactionally)
- Per-user data isolation: every board/list/card operation verifies
  ownership through the full hierarchy; cross-user access returns `404`
- Real-time sync: a `board_updated` WebSocket event is broadcast to every
  client subscribed to a board after any mutation, so multiple open tabs
  stay in sync automatically

## Running locally

### 1. Database

```sh
cd Golang/Web/TaskFlow
docker compose up -d
```

This starts PostgreSQL on `localhost:5433` (not the default 5432, to avoid
clashing with other local Postgres instances).

### 2. Apply migrations

```sh
cd backend
sh scripts/migrate.sh
```

### 3. Backend

```sh
cd backend
cp .env.example .env   # adjust if needed
go run ./cmd/api
```

The API listens on `http://localhost:8081` by default.

### 4. Frontend

```sh
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`. The Vite dev server proxies `/api/*` and
`/ws/*` to the backend, so no CORS configuration is needed in development.

## Running tests

```sh
cd backend
go test ./...
```

Unit tests cover the reorder/resequencing logic (`internal/service`) and
JWT generation/verification. They intentionally avoid requiring a live
database; a natural next step would be adding repository-level integration
tests with `testcontainers-go`.

## API overview

| Method | Path | Description |
|---|---|---|
| POST | `/auth/register` | Create an account, returns a JWT |
| POST | `/auth/login` | Sign in, returns a JWT |
| GET | `/auth/me` | Current user (requires auth) |
| GET / POST | `/boards` | List / create boards |
| GET / PUT / DELETE | `/boards/{boardID}` | Full board detail (nested lists+cards) / update / delete |
| POST | `/boards/{boardID}/lists` | Create a list |
| PUT / PATCH / DELETE | `/boards/{boardID}/lists/{listID}` | Update / **reorder** / delete a list |
| POST | `/boards/{boardID}/lists/{listID}/cards` | Create a card |
| PUT / PATCH / DELETE | `/boards/{boardID}/lists/{listID}/cards/{cardID}` | Update / **move** (cross-list capable) / delete a card |
| GET | `/ws/boards/{boardID}?token=<jwt>` | WebSocket subscription for live board updates |

All routes except `/auth/register`, `/auth/login`, and `/health` require a
`Authorization: Bearer <token>` header (the WebSocket route uses a `token`
query param instead, since browsers can't set custom headers during the
handshake).

## Possible next steps

- Board collaborators (share a board with other users, not just the owner)
- Repository-level integration tests with `testcontainers-go`
- Optimistic concurrency / conflict resolution for simultaneous edits
- Deployment setup (Dockerfile for the API, CI pipeline)
