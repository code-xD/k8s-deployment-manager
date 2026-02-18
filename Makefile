.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/api/main.go -o swagger --parseDependency --parseInternal
	@echo "Swagger docs generated in swagger/"

.PHONY: gen-orm
gen-orm:
	@echo "Generating GORM query code..."
	@go run cmd/gen-orm/main.go
	@echo "GORM query code generated in internal/database/query/"

.PHONY: run run-worker run-watcher

run:
	@go run cmd/api/main.go

run-worker:
	@go run cmd/worker/consumer/main.go

run-watcher:
	@go run cmd/worker/watcher/main.go

.PHONY: build build-watcher
build:
	@go build -o bin/api cmd/api/main.go

build-watcher:
	@go build -o bin/watcher cmd/worker/watcher/main.go

.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -f docker/api.dockerfile -t k8s-deployment-manager-api:latest .

.PHONY: docker-build-worker
docker-build-worker:
	@echo "Building worker Docker image..."
	@docker build -f docker/worker.dockerfile -t k8s-deployment-manager-worker:latest .

.PHONY: docker-build-watcher
docker-build-watcher:
	@echo "Building watcher Docker image..."
	@docker build -f docker/watcher.dockerfile -t k8s-deployment-manager-watcher:latest .

.PHONY: docker-build-all
docker-build-all: docker-build docker-build-worker docker-build-watcher
	@echo "All Docker images built successfully!"

.PHONY: docker-push
docker-push:
	@echo "Pushing Docker image..."
	@docker push k8s-deployment-manager-api:latest

.PHONY: docker-push-worker
docker-push-worker:
	@echo "Pushing worker Docker image..."
	@docker push k8s-deployment-manager-worker:latest

.PHONY: docker-push-watcher
docker-push-watcher:
	@echo "Pushing watcher Docker image..."
	@docker push k8s-deployment-manager-watcher:latest

.PHONY: k8s-namespace
k8s-namespace:
	@echo "Creating namespace if it doesn't exist..."
	@kubectl create namespace dep-manager --dry-run=client -o yaml | kubectl apply -f -

.PHONY: k8s-deploy
k8s-deploy: k8s-namespace
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f k8s/api/ -n dep-manager

.PHONY: k8s-deploy-worker
k8s-deploy-worker: k8s-namespace
	@echo "Deploying worker to Kubernetes..."
	@kubectl apply -f k8s/worker/clusterrole.yaml
	@kubectl apply -f k8s/worker/clusterrolebinding.yaml
	@kubectl apply -f k8s/worker/ -n dep-manager

.PHONY: k8s-deploy-watcher
k8s-deploy-watcher: k8s-namespace
	@echo "Deploying watcher to Kubernetes..."
	@kubectl apply -f k8s/watcher/clusterrole.yaml
	@kubectl apply -f k8s/watcher/clusterrolebinding.yaml
	@kubectl apply -f k8s/watcher/ -n dep-manager

.PHONY: k8s-deploy-all
k8s-deploy-all: k8s-deploy k8s-deploy-worker k8s-deploy-watcher
	@echo "All components deployed to Kubernetes successfully!"

.PHONY: build-and-deploy
build-and-deploy: docker-build-all k8s-deploy-all
	@echo "Build and deployment completed successfully!"

.PHONY: k8s-delete
k8s-delete:
	@echo "Deleting Kubernetes resources..."
	@kubectl delete -f k8s/api/ -n dep-manager

.PHONY: k8s-delete-worker
k8s-delete-worker:
	@echo "Deleting worker Kubernetes resources..."
	@kubectl delete -f k8s/worker/ -n dep-manager

.PHONY: k8s-delete-watcher
k8s-delete-watcher:
	@echo "Deleting watcher Kubernetes resources..."
	@kubectl delete -f k8s/watcher/ -n dep-manager
