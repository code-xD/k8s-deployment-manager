# internal/service

Business logic (use-cases) for the application.

- Implement interfaces from **pkg/ports** here.
- No HTTP, no message queues, no DB details—only domain rules and orchestration.
- Both **API handlers** (internal/api) and **worker handlers** (internal/worker) call these services.

Add one package per aggregate or bounded context if the app grows (e.g. service/deployment, service/health).
