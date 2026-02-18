# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the worker binary (no Swagger/GORM gen needed for consumer)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker ./cmd/worker/consumer/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/worker .

# Copy config folder
COPY --from=builder /app/config ./config

# Copy templates folder (needed for DeploymentManager to create K8s deployments)
COPY --from=builder /app/templates ./templates

# Run the binary
CMD ["./worker"]
