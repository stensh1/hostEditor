FROM golang:alpine

RUN apk add --no-cache git

RUN mkdir /app
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

RUN apk add --no-cache bash

# client files to build app
COPY . .
RUN rm -rf .bin cmd/server docker logs .gitignore dump.rdb Makefile README.md

# building
RUN go build -o client ./cmd/client

CMD ["bash"]