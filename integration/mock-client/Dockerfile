FROM golang:alpine

ADD . /go/src/github.com/smartcontractkit/external-initiator/integration/mock-client
RUN cd /go/src/github.com/smartcontractkit/external-initiator/integration/mock-client && go get && go build

EXPOSE 8080

ENTRYPOINT ["mock-client"]
