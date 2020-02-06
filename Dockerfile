FROM golang:alpine

ADD . /go/src/github.com/smartcontractkit/external-initiator
RUN cd /go/src/github.com/smartcontractkit/external-initiator && go get && go build

EXPOSE 8080

ENTRYPOINT ["external-initiator"]
