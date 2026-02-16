# API Deployment Guide

## Prerequisites
- Docker installed and running
- Kubernetes cluster running (minikube, kind, or cloud cluster)
- kubectl configured to connect to your cluster

## Quick Start

### 1. Build the Docker Image

**Using Makefile:**
```bash
make docker-build
```

**Or manually:**
```bash
docker build -f docker/api.dockerfile -t k8s-deployment-manager-api:latest .
```

### 2. Load Image into Kubernetes (for local clusters)

**For Minikube:**
```bash
minikube image load k8s-deployment-manager-api:latest
```

**For Kind:**
```bash
kind load docker-image k8s-deployment-manager-api:latest
```

**For Docker Desktop Kubernetes:**
The image should be available automatically if Docker Desktop is running.

### 3. Deploy to Kubernetes

**Using Makefile:**
```bash
make k8s-deploy
```

**Or manually:**
```bash
kubectl apply -f k8s/api/
```

This will create:
- ConfigMap (`api-config`)
- Deployment (`api`) with 2 replicas
- Service (`api`)

### 4. Verify Deployment

```bash
# Check pods
kubectl get pods -l app=api

# Check service
kubectl get svc api

# Check deployment
kubectl get deployment api

# View logs
kubectl logs -l app=api -f
```

### 5. Access the API

**Port Forward (for local testing):**
```bash
kubectl port-forward svc/api 8080:80
```

Then access:
- API: http://localhost:8080/api/v1/ping
- Swagger UI: http://localhost:8080/swagger/index.html

**Inside the cluster:**
- Service URL: `http://api:80/api/v1/ping`
- Swagger UI: `http://api:80/swagger/index.html`

## Production Deployment

### 1. Build and Push to Container Registry

```bash
# Tag for your registry
docker tag k8s-deployment-manager-api:latest YOUR_REGISTRY/k8s-deployment-manager-api:latest

# Push to registry
docker push YOUR_REGISTRY/k8s-deployment-manager-api:latest
```

### 2. Update Deployment Image

Edit `k8s/api/deployment.yaml` and change:
```yaml
image: YOUR_REGISTRY/k8s-deployment-manager-api:latest
imagePullPolicy: Always  # Change from IfNotPresent
```

### 3. Deploy

```bash
kubectl apply -f k8s/api/
```

## Useful Commands

**Delete deployment:**
```bash
make k8s-delete
# or
kubectl delete -f k8s/api/
```

**Scale replicas:**
```bash
kubectl scale deployment api --replicas=3
```

**Update deployment (after image change):**
```bash
kubectl rollout restart deployment api
```

**View deployment status:**
```bash
kubectl rollout status deployment api
```

**Get service URL:**
```bash
kubectl get svc api
```

## Troubleshooting

**Pods not starting:**
```bash
kubectl describe pod -l app=api
kubectl logs -l app=api
```

**Image pull errors:**
- For local: Ensure image is loaded into cluster
- For registry: Check imagePullSecrets and registry credentials

**Service not accessible:**
```bash
kubectl get endpoints api
kubectl describe svc api
```
