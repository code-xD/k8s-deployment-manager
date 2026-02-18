# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the watcher binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o watcher ./cmd/worker/watcher/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/watcher .

# Copy config folder
COPY --from=builder /app/config ./config

# Run the binary
CMD ["./watcher"]
