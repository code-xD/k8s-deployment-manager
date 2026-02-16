# Kubernetes Manifests

Manifests for deploying the stack: **Postgres**, **NATS**, **API** (Go), and **Worker** (Go).

## Repo layout (recommended)

```
k8s-deployment-manager/
├── cmd/
│   ├── api/          # Go API server (main package)
│   └── worker/       # Go worker (main package)
├── internal/         # Shared Go code (DB, NATS client, etc.)
├── k8s/              # All Kubernetes manifests (this folder)
│   ├── postgres/
│   ├── nats/
│   ├── api/
│   └── worker/
├── go.mod
├── go.sum
└── README.md
```

- **Infrastructure first:** Postgres and NATS are deployed so the API and worker can connect.
- **App later:** API and worker manifests reference the same secrets/config and image names you’ll use when the Go code is built.

## Approach

1. **Order of deployment:** Postgres → NATS → API → Worker (API and worker depend on DB and NATS).
2. **Secrets:** Passwords and sensitive config live in Secrets; use placeholders in repo and inject via CI or a secret manager in prod.
3. **Config:** Non-sensitive config in ConfigMaps (DB host, NATS URL, feature flags).
4. **Images:** API and worker will use your registry (e.g. `your-registry/api:tag`, `your-registry/worker:tag`). Build and push in CI.
5. **API/worker env:** Use `postgres-secret` for DB; set `NATS_URL=nats://nats:4222`.

## What’s in this folder

| Directory   | Purpose                          | Deploy when        |
|------------|-----------------------------------|--------------------|
| `postgres/`| PostgreSQL (StatefulSet, PVC)     | First (data store) |
| `nats/`    | NATS (Deployment)                 | Second (messaging) |
| `api/`     | Go API deployment (placeholder)    | After API code     |
| `worker/`  | Go Worker deployment (placeholder)| After worker code  |

## Deploy (local / dev)

From repo root:

```bash
# 1. Infrastructure
kubectl apply -f k8s/postgres/
kubectl apply -f k8s/nats/

# Wait for Postgres and NATS to be ready, then:
# kubectl apply -f k8s/api/
# kubectl apply -f k8s/worker/
```

## Namespace

Manifests use the `default` namespace. For a dedicated namespace (e.g. `app`), create it and apply with:

```bash
kubectl create namespace app
kubectl apply -f k8s/postgres/ -n app
kubectl apply -f k8s/nats/ -n app
```

Or add `namespace: app` to each manifest / use Kustomize overlays later.
