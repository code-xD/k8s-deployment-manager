# Setup

This document provides instructions for setting up and deploying the K8s Deployment Manager system.

## Prerequisites

- Kubernetes cluster (local or remote)
- `kubectl` configured to access your cluster
- Docker (for building images)
- Go 1.26+ (for local development)

## Deployment Steps

### 1. Deploy Third-Party Components

Deploy PostgreSQL and NATS in order:

```bash
# Deploy PostgreSQL
kubectl apply -f k8s/postgres/

# Deploy NATS
kubectl apply -f k8s/nats/
```

**Note:** Wait for both services to be ready before proceeding to the next step.

### 2. Deploy Application Components

Deploy API, Worker, and Watcher:

```bash
# Deploy API
kubectl apply -f k8s/api/

# Deploy Worker (includes ClusterRole and ClusterRoleBinding)
kubectl apply -f k8s/worker/clusterrole.yaml
kubectl apply -f k8s/worker/clusterrolebinding.yaml
kubectl apply -f k8s/worker/

# Deploy Watcher (includes ClusterRole and ClusterRoleBinding)
kubectl apply -f k8s/watcher/clusterrole.yaml
kubectl apply -f k8s/watcher/clusterrolebinding.yaml
kubectl apply -f k8s/watcher/
```

### 3. Verify Deployment

Check that all pods are running:

```bash
kubectl get pods -n dep-manager
```

## Local Development Setup

### Port Forwarding

For local testing, port-forward PostgreSQL and NATS:

```bash
# Port forward PostgreSQL (default port 5432)
kubectl port-forward svc/postgres 5432:5432 -n dep-manager

# Port forward NATS (default port 4222)
kubectl port-forward svc/nats 4222:4222 -n dep-manager
```

### Run Locally

In separate terminals, run each component:

```bash
# Terminal 1: Run API
make run

# Terminal 2: Run Worker
make run-worker

# Terminal 3: Run Watcher
make run-watcher
```

## Production Deployment

### Build Docker Images

Build all Docker images:

```bash
# Build all images
make docker-build-all

# Or build individually
make docker-build          # API image
make docker-build-worker   # Worker image
make docker-build-watcher  # Watcher image
```

### Push Images (if using remote registry)

```bash
# Push all images
make docker-push docker-push-worker docker-push-watcher

# Or push individually
make docker-push
make docker-push-worker
make docker-push-watcher
```

**Note:** Update image names in Kubernetes manifests if using a custom registry.

### Deploy to Kubernetes

Deploy all components:

```bash
# Deploy everything (creates namespace, applies all manifests)
make build-and-deploy

# Or deploy individually
make k8s-deploy          # API
make k8s-deploy-worker   # Worker
make k8s-deploy-watcher  # Watcher
```

### Using Makefile Commands

The Makefile provides convenient commands:

```bash
# Build and deploy everything
make build-and-deploy

# Deploy all Kubernetes resources
make k8s-deploy-all

# Delete deployments
make k8s-delete          # API
make k8s-delete-worker   # Worker
make k8s-delete-watcher  # Watcher
```

## Configuration

### Environment Variables

Components are configured via ConfigMaps and environment variables. Key configuration includes:

- **Database**: Connection details for PostgreSQL
- **NATS**: Connection details and stream configuration
- **Kubernetes**: Cluster configuration and manager tag for filtering deployments

### Secrets

PostgreSQL credentials are stored in a Kubernetes Secret:

```bash
# Create secret (example)
kubectl create secret generic postgres-secret \
  --from-literal=POSTGRES_USER=app \
  --from-literal=POSTGRES_PASSWORD=your-password \
  --from-literal=POSTGRES_DB=appdb \
  -n dep-manager
```

**Security Note:** In production, use a proper secret management solution (e.g., External Secrets Operator, Vault).

## Namespace

By default, all resources are deployed to the `dep-manager` namespace. The Makefile automatically creates this namespace if it doesn't exist.

To use a different namespace, update the manifests or override with:

```bash
kubectl apply -f k8s/api/ -n your-namespace
```

## Troubleshooting

### Check Pod Logs

```bash
# API logs
kubectl logs -f deployment/api -n dep-manager

# Worker logs
kubectl logs -f deployment/worker -n dep-manager

# Watcher logs
kubectl logs -f deployment/watcher -n dep-manager
```

### Check Service Status

```bash
# Check all services
kubectl get svc -n dep-manager

# Check specific service
kubectl describe svc api -n dep-manager
```

### Database Connection Issues

Verify PostgreSQL is accessible:

```bash
# Port forward and test connection
kubectl port-forward svc/postgres 5432:5432 -n dep-manager
psql -h localhost -U app -d appdb
```

### NATS Connection Issues

Verify NATS is accessible:

```bash
# Port forward and test connection
kubectl port-forward svc/nats 4222:4222 -n dep-manager
# Use NATS CLI or client library to test
```

## Next Steps

After setup, refer to:
- [Architecture Documentation](./architecture.md) for system overview
- [Design Decisions](./design-decisions.md) for architectural rationale
- [Schema Documentation](./schema.md) for database structure
- API documentation (Swagger) at `/swagger/index.html` when API is running
