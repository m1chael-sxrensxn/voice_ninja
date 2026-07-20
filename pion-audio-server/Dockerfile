# Build stage
FROM golang:1.26 AS builder

WORKDIR /app

# Download dependencies first (better layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build a static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o server ./cmd/server

# Runtime stage
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/server /server
COPY static/index.html /static/index.html
COPY static/index.js /static/index.js

EXPOSE 8080

ENTRYPOINT ["/server"]
