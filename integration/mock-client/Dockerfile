FROM golang:alpine as build-env

RUN apk add build-base linux-headers

# We copy the complete local /external-initiator, as it's sometimes useful during
# development to replace the external-initiator go dependency with a local version.
RUN mkdir -p /external-initiator
WORKDIR /external-initiator
COPY . .
WORKDIR /external-initiator/integration/mock-client
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/mock-client

FROM scratch
COPY --from=build-env /go/bin/mock-client /go/bin/mock-client
COPY --from=build-env /external-initiator/integration/mock-client/blockchain/mock-responses /blockchain/mock-responses

EXPOSE 8080

ENTRYPOINT ["/go/bin/mock-client"]
