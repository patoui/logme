FROM golang:1.18.2-alpine3.16

COPY go.mod go.sum /go/src/github.com/patoui/logme/
WORKDIR /go/src/github.com/patoui/logme

RUN go mod download
RUN apk add --no-cache git

COPY . .

RUN apk add --no-cache ca-certificates && update-ca-certificates

RUN go install github.com/githubnemo/CompileDaemon@latest

EXPOSE 8080 8080

ENTRYPOINT CompileDaemon -build='go build -o build/logme -buildvcs=false github.com/patoui/logme' -command='./build/logme' -color=true 2>&1