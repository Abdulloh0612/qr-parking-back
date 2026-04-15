FROM golang:1.24 AS local

WORKDIR /usr/src/app

RUN go install github.com/cosmtrek/air@v1.51.0 && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0

COPY go.mod go.sum ./
RUN go mod download

RUN apt-get update && apt-get install -y sshpass

COPY . .

CMD ["air"]

FROM golang:1.24 AS builder

WORKDIR /usr/src/app

# Install migrate tool
RUN CGO_ENABLED=0 GOOS=linux go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0

# Copy and download dependencies first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Create necessary directories
RUN mkdir -p /usr/src/app/tmp /usr/src/app/cmd

# Copy the rest of the source code
COPY . .

# Build server executable
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /usr/src/app/tmp/server_process ./main.go

# Build command executables
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/src/app/cmd/generate_tokens ./cmd/generate_tokens/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /usr/src/app/cmd/migrate ./cmd/migrate/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /usr/src/app/cmd/seed ./cmd/seed/main.go

FROM alpine:3.19 AS production

WORKDIR /usr/bin/app

# Install required packages and create directories in a single layer
RUN apk add --no-cache bash curl && \
    mkdir -p /usr/bin/app/cmd/logger

RUN apk add --no-cache bash curl sshpass
RUN apk add --no-cache openssl

# Copy configuration files
COPY .env /usr/bin/app/.env
COPY docker/app/health-check.sh /usr/bin/app/health-check.sh
RUN ln -s /usr/bin/app/.env /usr/bin/app/cmd/.env

# Copy server executable and migrate tool
COPY --from=builder --chmod=755 /usr/src/app/tmp/server_process /usr/bin/app/server_process
COPY --from=builder --chmod=755 /go/bin/migrate /usr/bin/migrate

# Copy command executables
COPY --from=builder --chmod=755 /usr/src/app/cmd/generate_tokens /usr/bin/app/cmd/generate_tokens
COPY --from=builder --chmod=755 /usr/src/app/cmd/migrate /usr/bin/app/cmd/migrate
COPY --from=builder --chmod=755 /usr/src/app/cmd/seed /usr/bin/app/cmd/seed

CMD ["/usr/bin/app/health-check.sh"]
