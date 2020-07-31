FROM golang:alpine as build-env

RUN apk add build-base linux-headers

# NOTICE: We copy the complete local /external-initiator, as it's sometimes useful during
# development to replace the external-initiator go dependency with a local version.

# First, we copy go mod files to cache Docker image layers and avoid running
# 'go mod download' if dependecies did not change.
RUN mkdir -p /external-initiator/integration/mock-client
WORKDIR /external-initiator
COPY go.mod go.sum ./
WORKDIR /external-initiator/integration/mock-client
COPY /integration/mock-client/go.mod /integration/mock-client/go.sum ./
RUN go mod download

# Second, we copy (including the /external-initiator as a local dependency) and compile the code
WORKDIR /external-initiator
COPY . .
WORKDIR /external-initiator/integration/mock-client
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/mock-client

FROM scratch
COPY --from=build-env /go/bin/mock-client /go/bin/mock-client
COPY --from=build-env /external-initiator/integration/mock-client/blockchain/mock-responses /blockchain/mock-responses

EXPOSE 8080

ENTRYPOINT ["/go/bin/mock-client"]
