FROM golang:1.15-alpine as build-env

RUN apk add build-base linux-headers

# First, we copy go mod files to cache Docker image layers and avoid running
# 'go mod download' if dependecies did not change.
RUN mkdir -p /external-initiator/integration/mock-client
WORKDIR /external-initiator
COPY go.mod go.sum ./
RUN go mod download

# Second, we copy (including the /external-initiator as a local dependency) and compile the code
WORKDIR /external-initiator
COPY . .
WORKDIR /external-initiator/integration/mock-client
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/mock-client

FROM alpine:latest
COPY --from=build-env /go/bin/mock-client /go/bin/mock-client
COPY --from=build-env /external-initiator/integration/mock-client/blockchain/static /blockchain/static

EXPOSE 8080
EXPOSE 8090

ENTRYPOINT ["/go/bin/mock-client"]
