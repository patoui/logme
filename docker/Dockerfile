FROM golang:1.22.2-alpine3.19

COPY go.mod go.sum /go/src/github.com/patoui/logme/
WORKDIR /go/src/github.com/patoui/logme

RUN go mod download
RUN apk add git
RUN apk add gcc
RUN apk add libc-dev

COPY . .

RUN apk add --no-cache ca-certificates && update-ca-certificates

RUN go install github.com/githubnemo/CompileDaemon@latest

EXPOSE 8080 8080

ENTRYPOINT CompileDaemon -build='go build -o build/logme -buildvcs=false github.com/patoui/logme/cmd/logme' -command='./build/logme' -color=true 2>&1