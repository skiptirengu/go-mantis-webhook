FROM golang:latest

WORKDIR /go/src/github.com/skiptirengu/go-mantis-webhook
ADD . .
RUN go get -v ./
RUN go build

CMD ["go-mantis-webhook"]