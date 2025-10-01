# Stage 1: Build
FROM golang:1.25-alpine AS builder

# Install git if needed (for go get)
RUN apk add --no-cache git

WORKDIR /src

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build a static coordinator binary from the project cmd directory
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /app/coordinator ./cmd/coordinator

# Stage 2: Run
FROM alpine:latest

WORKDIR /app
COPY ./config /app/config
COPY --from=builder /app/coordinator .

# Coordinator listens on 50051 in the coordinator server implementation
EXPOSE 50051

# Run the coordinator binary
CMD ["./coordinator"]
