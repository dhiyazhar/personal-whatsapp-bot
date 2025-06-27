FROM golang:1.24-alpine AS BUILDER

RUN apk add --no-cache git ca-certificates 

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY src/ ./src/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /whatsapp-bot ./src

FROM python:3.11-slim-bookworm

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg && \
    rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir -U yt-dlp

COPY --from=builder /whatsapp-bot .

EXPOSE 8080

CMD ["./whatsapp-bot"]