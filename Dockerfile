FROM golang:1.25-alpine AS builder
RUN apk update && apk upgrade && apk add --no-cache bash git openssh
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .


FROM jauderho/yt-dlp:2025.10.22
WORKDIR /app
COPY --from=builder /app/main /app/main
COPY --from=builder /app/config.ini /app/config.ini
COPY --from=builder /app/YTM_cookies.txt /app/YTM_cookies.txt

ENTRYPOINT []
CMD ["/app/main"]
