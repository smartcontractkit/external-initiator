FROM golang:alpine as build-env

RUN apk add build-base linux-headers
RUN mkdir /mock-client
WORKDIR /mock-client
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/mock-client

FROM scratch
COPY --from=build-env /go/bin/mock-client /go/bin/mock-client
COPY --from=build-env /mock-client/blockchain/mock-responses /blockchain/mock-responses

EXPOSE 8080

ENTRYPOINT ["/go/bin/mock-client"]
