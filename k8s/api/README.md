# API Deployment Guide

## Prerequisites
- Docker installed and running
- Kubernetes cluster running (minikube, kind, or cloud cluster)
- kubectl configured to connect to your cluster
- Namespace `dep-manager` already created

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
kubectl apply -f k8s/api/ -n dep-manager
```

This will create in the `dep-manager` namespace:
- ConfigMap (`api-config`)
- Deployment (`api`) with 2 replicas
- Service (`api`)

### 4. Verify Deployment

```bash
# Check pods
kubectl get pods -l app=api -n dep-manager

# Check service
kubectl get svc api -n dep-manager

# Check deployment
kubectl get deployment api -n dep-manager

# View logs
kubectl logs -l app=api -f -n dep-manager
```

### 5. Access the API

**Port Forward (for local testing):**
```bash
kubectl port-forward svc/api 8080:80 -n dep-manager
```

Then access:
- API: http://localhost:8080/api/v1/ping
- Swagger UI: http://localhost:8080/swagger/index.html

**Inside the cluster:**
- Service URL (same namespace): `http://api:80/api/v1/ping`
- Service URL (cross namespace): `http://api.dep-manager.svc.cluster.local:80/api/v1/ping`
- Swagger UI: `http://api:80/swagger/index.html` or `http://api.dep-manager.svc.cluster.local:80/swagger/index.html`

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
kubectl apply -f k8s/api/ -n dep-manager
```

## Useful Commands

**Set default namespace (optional):**
```bash
kubectl config set-context --current --namespace=dep-manager
```

**Delete deployment:**
```bash
make k8s-delete
# or
kubectl delete -f k8s/api/ -n dep-manager
```

**Scale replicas:**
```bash
kubectl scale deployment api --replicas=3 -n dep-manager
```

**Update deployment (after image change):**
```bash
kubectl rollout restart deployment api -n dep-manager
```

**View deployment status:**
```bash
kubectl rollout status deployment api -n dep-manager
```

**Get service URL:**
```bash
kubectl get svc api -n dep-manager
```

**List all resources in namespace:**
```bash
kubectl get all -n dep-manager
```

## Troubleshooting

**Pods not starting:**
```bash
kubectl describe pod -l app=api -n dep-manager
kubectl logs -l app=api -n dep-manager
```

**Image pull errors (ImagePullBackOff):**

1. **Build the Docker image first:**
   ```bash
   make docker-build
   ```

2. **Load image into your Kubernetes cluster:**

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

3. **Restart the deployment:**
   ```bash
   kubectl rollout restart deployment api -n dep-manager
   ```

4. **Check image exists:**
   ```bash
   docker images | grep k8s-deployment-manager-api
   ```

5. **For production/registry:**
   - Tag and push image to your registry
   - Update deployment.yaml with registry URL
   - Set imagePullPolicy to Always
   - Configure imagePullSecrets if needed

**Service not accessible:**
```bash
kubectl get endpoints api -n dep-manager
kubectl describe svc api -n dep-manager
```

**Check namespace:**
```bash
# List all namespaces
kubectl get namespaces

# Check current context namespace
kubectl config view --minify | grep namespace

# List resources in namespace
kubectl get all -n dep-manager
```
