FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application pointing to the main file
RUN go build -o main cmd/main.go

FROM alpine:latest

WORKDIR /root/

# Install CA certificates for SSL connections
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .
COPY --from=builder /app/assets ./assets

EXPOSE 8080

CMD ["./main"]
