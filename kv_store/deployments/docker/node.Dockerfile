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

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /app/node ./cmd/node

# Stage 2: Run
FROM alpine:latest


WORKDIR /app
COPY --from=builder /app/node .

# Run the binary
CMD ["./node"]
