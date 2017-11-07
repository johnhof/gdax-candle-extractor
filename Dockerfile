FROM golang:1.9-alpine3.6

WORKDIR /go/src/extractor

COPY . .

RUN apk update && \
    apk add git && \
    go get -u github.com/golang/dep/cmd/dep && \
    dep ensure && \
    go build -o extract main.go


CMD ["/go/src/extractor/extract"]