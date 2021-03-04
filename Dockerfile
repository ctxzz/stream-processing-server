FROM golang:latest
RUN mkdir -p /go/src
WORKDIR /go/src
ADD . /go/src
