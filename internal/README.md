# internal

Private application code. Only the main modules under `cmd/` (api, worker, etc.) and other packages under `internal/` may import from here.

## Layout

- **service/** — Business logic (use-cases). Single source of truth for rules and workflows. Used by both API and worker. Depends on **pkg/ports** only (e.g. repository interfaces).
- **repository/** — DB access: implements **pkg/ports** (e.g. `DeploymentRepository`). Contains **postgres/** (and optionally redis/, etc.). Used by service via dependency injection.
- **api/** — API-specific glue: HTTP handlers, middleware, route wiring. Handlers call `service` only; no business logic here.
- **worker/** — Worker-specific glue: message/job consumers, schedulers. Call `service` only; no business logic here.

## Flow

```
HTTP request  →  cmd/api  →  internal/api (handler)  →  internal/service  →  internal/repository (postgres)
Message/Job   →  cmd/worker  →  internal/worker (handler)  →  internal/service  →  internal/repository (postgres)
```

Business logic lives only in `internal/service`. Repo functions that call the database live in `internal/repository` and implement interfaces from `pkg/ports`.
