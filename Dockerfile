FROM golang:latest
RUN mkdir -p /go/src
WORKDIR /go/src
ADD . /go/src

RUN go get -u go.mongodb.org/mongo-driver/mongo \
 && go get -u github.com/codegangsta/gin
