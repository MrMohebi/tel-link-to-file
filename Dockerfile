FROM golang:1.18.3-alpine as builder
RUN apk update && apk upgrade && apk add --no-cache bash git openssh
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .


FROM python:alpine3.16
RUN pip install spotdl
RUN spotdl --download-ffmpeg
COPY --from=builder /app/main /
COPY --from=builder /app/config.ini /
CMD ["./main"]