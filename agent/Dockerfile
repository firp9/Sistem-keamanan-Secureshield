# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

RUN apk add --no-cache libpcap-dev gcc musl-dev

COPY go.mod go.sum ./  
RUN go mod download

COPY . .

RUN go build -o agent .

# Final stage
FROM alpine:latest

RUN apk add --no-cache libpcap

WORKDIR /app

COPY --from=builder /app/agent .

CMD ["./agent"]
