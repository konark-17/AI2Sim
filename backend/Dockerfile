# ─── Stage 1: Build Go binary ───────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Cache module downloads
COPY go.mod go.sum ./
RUN go mod download

# Build the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o ai2sim-server ./cmd/server

# ─── Stage 2: Ollama + server runtime ───────────────────────────────────────
FROM ubuntu:24.04

# Install Ollama
RUN apt-get update && apt-get install -y curl ca-certificates && \
    curl -fsSL https://ollama.com/install.sh | sh && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Copy Go binary
COPY --from=builder /app/ai2sim-server /usr/local/bin/ai2sim-server

# Copy entrypoint script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Ollama data directory
VOLUME ["/root/.ollama"]

# Expose API port
EXPOSE 8080

ENV PORT=8080
ENV OLLAMA_URL=http://localhost:11434
ENV OLLAMA_MODEL=qwen2.5-coder:7b
ENV ALLOWED_ORIGINS=*

ENTRYPOINT ["/docker-entrypoint.sh"]
