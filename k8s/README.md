# Kubernetes Manifests

Kubernetes deployment manifests for the stack.

## Setup

Deploy in order: Postgres → NATS → API → Worker

```bash
kubectl apply -f k8s/postgres/
kubectl apply -f k8s/nats/
kubectl apply -f k8s/api/
kubectl apply -f k8s/worker/
```

## Structure

- `postgres/` — PostgreSQL StatefulSet
- `nats/` — NATS with JetStream
- `api/` — API server deployment
- `worker/` — Worker deployment

## Configuration

- Secrets: Use `postgres-secret` for DB credentials
- ConfigMaps: Non-sensitive config (DB host, NATS URL)
- Namespace: Defaults to `default`, can be overridden
