# internal

Private application code. Only `cmd/` modules may import from here.

## Architecture

- **service/** — Business logic. Implements interfaces from `pkg/ports/service`
- **repository/** — Data access. Implements interfaces from `pkg/ports/repo`
- **api/** — HTTP handlers and middleware. Calls services only
- **worker/** — Message/job consumers. Calls services only

## Conventions

- Services depend on `pkg/ports` interfaces only
- Repositories implement `pkg/ports/repo` interfaces
- No business logic in handlers (api/worker)
- Wire dependencies in `cmd/` composition root
