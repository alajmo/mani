FROM alpine:3.10

COPY --from=golang:1.16.3-alpine /usr/local/go/ /usr/local/go/

ENV XDG_CACHE_HOME=/tmp/.cache
ENV GOPATH=${HOME}/go
ENV GO111MODULE=on
ENV PATH="/usr/local/go/bin:${PATH}"

RUN apk update
RUN apk add --no-cache make build-base bash curl g++ git

RUN addgroup -g 1000 -S test && adduser -u 1000 -S test -G test

WORKDIR /home/test

COPY . .
COPY ./test/git /usr/local/bin/git

RUN go mod download && make build && cp /home/test/execs/mani /usr/local/bin/mani

USER test
