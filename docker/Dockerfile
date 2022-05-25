FROM golang:1.18.2-alpine3.16

COPY go.mod go.sum /go/src/github.com/patoui/logme/
WORKDIR /go/src/github.com/patoui/logme

RUN go mod download

COPY . .

RUN apk add --no-cache ca-certificates && update-ca-certificates
RUN apk add --update npm

ENV GO111MODULE=on
RUN go get github.com/githubnemo/CompileDaemon

EXPOSE 8080 8080

ENTRYPOINT CompileDaemon -build='go build -o build/logme github.com/patoui/logme' -command='./build/logme' -color=true 2>&1
