FROM alpine:3.21.0

ENV GOCACHE=/go/cache
ENV GO111MODULE=on
ENV PATH="/usr/local/go/bin:${PATH}"
ENV USER="test"
ENV HOME="/home/test"

COPY --from=golang:1.25.5-alpine /usr/local/go/ /usr/local/go/

RUN apk update
RUN apk add --no-cache make build-base bash curl g++ git

RUN addgroup -g 1000 -S test && adduser -u 1000 -S test -G test

WORKDIR /home/test

COPY --chown=test go.mod go.sum ./
RUN go mod download
COPY --chown=test . .
COPY --chown=test ./test/scripts/git /usr/local/bin/git
RUN make build-test && cp /home/test/dist/mani /usr/local/bin/mani

USER test
