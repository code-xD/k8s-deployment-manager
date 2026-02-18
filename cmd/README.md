# cmd

Application entry points (composition root).

## Structure

- **api/** — HTTP API server
- **worker/** — Background worker

## Conventions

- Only place where concrete implementations are instantiated
- All dependencies wired here via dependency injection
- Other layers interact via interfaces from `pkg/ports/`
