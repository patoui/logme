FROM golang:1.22.2-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy && go build ./cmd/queue

FROM alpine:latest

RUN apk --no-cache add supervisor

COPY --from=builder /app/queue /usr/local/bin/queue

COPY ./.env /home/.env

RUN mkdir -p /etc/supervisor/conf.d

COPY ./docker/queue/supervisor.conf /etc/supervisor/supervisor.conf

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisor.conf"]
