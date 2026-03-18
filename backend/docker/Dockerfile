FROM golang:1.25 AS builder

WORKDIR /app

# 下载依赖项
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 编译 app
RUN go build -o main cmd/main.go
# 编译 worker
RUN go build -o worker cmd/worker/main.go


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
COPY --from=builder /app/worker .
RUN mkdir -p /assets

CMD ["./main"]
