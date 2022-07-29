FROM golang:1.17-alpine as build

WORKDIR /
COPY . .

RUN go build

FROM alpine

COPY --from=build /couchdb-bootstrapper /couchdb-bootstrapper

CMD ["/couchdb-bootstrapper", "run"]