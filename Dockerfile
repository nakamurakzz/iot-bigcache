# syntax=docker/dockerfile:1

FROM golang:1.20.6-alpine3.18

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY main.go ./

RUN go build -o /main

CMD [ "/main" ]