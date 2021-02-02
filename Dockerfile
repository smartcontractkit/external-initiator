FROM golang:1.15-alpine as build-env

RUN apk add build-base linux-headers
RUN apk --update add ca-certificates

RUN mkdir /external-initiator
WORKDIR /external-initiator
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Delete ./integration folder that is not needed in the context of external-initiator,
# but is required in the context of mock-client build.
RUN rm -rf ./integration
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/external-initiator

FROM alpine:latest

COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-env /go/bin/external-initiator /go/bin/external-initiator

EXPOSE 8080

ENTRYPOINT ["/go/bin/external-initiator"]
