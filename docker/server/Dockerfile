FROM golang:alpine

# Install sudo
RUN apk update && apk add sudo

# New user
RUN adduser -D user && echo "user:admin" | chpasswd

# New user in sudoers
RUN echo "user ALL=(ALL) ALL" >> /etc/sudoers

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

# server files to build app
COPY .. .
RUN rm -rf .bin cmd/client docker logs .gitignore dump.rdb Makefile README.md

# building
RUN go build -o server ./cmd/server

RUN chown -R user:user /app


USER user
CMD ["./server"]