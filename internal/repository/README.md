# internal/repository

Database repository implementations. They implement interfaces from **pkg/ports** and are used by **internal/service**.

- **postgres/** — PostgreSQL implementations (e.g. `DeploymentRepository`). Use `database/sql`, pgx, or an ORM here.
- Add other stores if needed (e.g. **redis/** for cache).

Service layer depends on ports (interfaces), not on this package directly; wire concrete repos in `cmd/api` / `cmd/worker`.
