FROM golang
WORKDIR /go/src/app

COPY ./go .

RUN go get gopkg.in/mgo.v2
RUN go build -o app

ENTRYPOINT ./app
