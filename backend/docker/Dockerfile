FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN go build -o main cmd/main.go

FROM debian:bookworm-slim

WORKDIR /root/

# Install CA certificates and document processing tools (OCR, Word support)
# tesseract-ocr: for OCR
# tesseract-ocr-chi-sim: Chinese simplified language pack for OCR
# antiword: for .doc extraction
# catdoc: fallback for .doc extraction
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tesseract-ocr \
    tesseract-ocr-chi-sim \
    antiword \
    catdoc \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/main .
RUN mkdir -p /assets

EXPOSE 8080

CMD ["./main"]
