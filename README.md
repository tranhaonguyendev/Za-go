# Zago

`Zago` is a Go library for working with Zalo in a cleaner, modular layout.

## Highlights

- Root package kept small and public-facing
- Internal implementation isolated under `internal/`
- Dedicated documentation website under `documents/`
- Go module based project layout

## Project Structure

```text
Zago/
├── documents/              # VitePress documentation website
├── internal/
│   ├── api/                # grouped API services
│   ├── app/                # session and state
│   ├── auth/               # authentication helpers
│   ├── core/               # shared domain primitives
│   ├── logger/             # logging helpers
│   ├── util/               # common utilities
│   └── worker/             # event/message objects
├── doc.go
├── go.mod
├── socket_callbacks.go
├── types.go
└── zalo.go
```

## Requirements

- Go `1.22+`
- Node.js `20+` for the documentation website

## Development

Compile the Go library:

```bash
go test ./...
```

Even without test files, `go test ./...` is still a convenient compile check for the whole module.

## Documentation Website

The docs site lives in `documents/` and runs on port `14711`.

Install dependencies:

```bash
cd documents
npm install
```

Run locally:

```bash
npm run docs:dev
```

Build static docs:

```bash
npm run docs:build
```

Preview the built site:

```bash
npm run docs:preview
```

## Public API Entry Points

- `Zalo(...)` creates a new client instance
- `SocketCallbacks` wires realtime event handlers
- `Message`, `User`, `Group`, `ThreadType` are re-exported for consumers

## Notes

- The current module path is `github.com/tranhaonguyendev/Za-go`
- If you publish this under a different repository path, update `go.mod`
