FROM golang:alpine as build-env

RUN apk add build-base linux-headers
RUN apk --update add ca-certificates

RUN mkdir /external-initiator
WORKDIR /external-initiator
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/external-initiator

FROM scratch

COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-env /go/bin/external-initiator /go/bin/external-initiator

EXPOSE 8080

ENTRYPOINT ["/go/bin/external-initiator"]
