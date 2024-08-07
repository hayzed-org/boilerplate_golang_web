FROM golang:1.23rc2-alpine3.20 AS build-stage

WORKDIR /

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o telex_be main.go

FROM alpine:latest

RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

WORKDIR /

COPY --from=build-stage telex_be .

EXPOSE 8080

USER nonroot:nonroot


ENTRYPOINT ["./telex_be"]
