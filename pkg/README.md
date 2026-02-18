# pkg

Shared packages used across the application.

## Structure

- **ports/** — Interface definitions (service and repo contracts)
- **dto/** — Data transfer objects and models
- **logger/** — Logging utilities
- **config/** — Configuration management
- **utils/** — Common utilities

## Conventions

- All interfaces defined in `ports/`
- Services implement `ports/service/*` interfaces
- Repositories implement `ports/repo/*` interfaces
- DTOs in `dto/` for API boundaries
