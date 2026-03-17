FROM golang:1.22-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -ldflags "-s -w" -o llmview .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/llmview /usr/local/bin/llmview
EXPOSE 4700
ENTRYPOINT ["llmview"]
