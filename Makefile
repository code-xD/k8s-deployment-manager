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

.PHONY: run
run:
	@go run cmd/api/main.go

.PHONY: build
build:
	@go build -o bin/api cmd/api/main.go

.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -f docker/api.dockerfile -t k8s-deployment-manager-api:latest .

.PHONY: docker-build-worker
docker-build-worker:
	@echo "Building worker Docker image..."
	@docker build -f docker/worker.dockerfile -t k8s-deployment-manager-worker:latest .

.PHONY: docker-push
docker-push:
	@echo "Pushing Docker image..."
	@docker push k8s-deployment-manager-api:latest

.PHONY: docker-push-worker
docker-push-worker:
	@echo "Pushing worker Docker image..."
	@docker push k8s-deployment-manager-worker:latest

.PHONY: k8s-deploy
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f k8s/api/ -n dep-manager

.PHONY: k8s-deploy-worker
k8s-deploy-worker:
	@echo "Deploying worker to Kubernetes..."
	@kubectl apply -f k8s/worker/ -n dep-manager

.PHONY: k8s-delete
k8s-delete:
	@echo "Deleting Kubernetes resources..."
	@kubectl delete -f k8s/api/ -n dep-manager

.PHONY: k8s-delete-worker
k8s-delete-worker:
	@echo "Deleting worker Kubernetes resources..."
	@kubectl delete -f k8s/worker/ -n dep-manager
