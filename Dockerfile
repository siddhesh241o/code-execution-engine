FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o engine ./cmd/server

FROM alpine:3.19
WORKDIR /root/
COPY --from=builder /app/engine .

EXPOSE 8080
CMD ["./engine"]