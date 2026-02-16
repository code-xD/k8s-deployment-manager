.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g app/api/main.go -o swagger --parseDependency --parseInternal
	@echo "Swagger docs generated in swagger/"

.PHONY: run
run:
	@go run app/api/main.go

.PHONY: build
build:
	@go build -o bin/api app/api/main.go
