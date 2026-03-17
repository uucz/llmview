# Stage 1: Build Svelte UI
FROM node:20-alpine AS ui
WORKDIR /app/ui
COPY ui/package*.json ./
RUN npm ci
COPY ui/ ./
RUN npm run build

# Stage 2: Build Go binary (pure Go, no CGo)
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui /app/ui/build internal/server/ui
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o llmview .

# Stage 3: Minimal runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/llmview /usr/local/bin/llmview
EXPOSE 4700
ENTRYPOINT ["llmview"]
