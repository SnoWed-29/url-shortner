# Build stage
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Download dependencies first for better Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the API
RUN CGO_ENABLED=0 GOOS=linux go build -o /url-shortener-api ./cmd/api


# Runtime stage
FROM alpine:3.22

WORKDIR /app

# CA certificates are needed for HTTPS requests
RUN apk --no-cache add ca-certificates

COPY --from=builder /url-shortener-api .

EXPOSE 8080

CMD ["./url-shortener-api"]