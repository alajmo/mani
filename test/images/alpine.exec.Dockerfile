FROM alpine:3.15.4 as build

ENV XDG_CACHE_HOME=/tmp/.cache
ENV GOPATH=${HOME}/go
ENV GO111MODULE=on
ENV PATH="/usr/local/go/bin:${PATH}"
ENV USER="test"
ENV HOME="/home/test"

COPY --from=golang:1.18.1-alpine /usr/local/go/ /usr/local/go/

RUN apk update
RUN apk add --no-cache make build-base bash curl g++ git

WORKDIR /opt

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build

FROM alpine:3.15.4

RUN apk update
RUN apk add --no-cache sudo bash zsh fish bash-completion git

COPY --from=build /opt/dist/mani /usr/local/bin/mani

RUN mani completion bash > /usr/share/bash-completion/completions/mani

RUN addgroup -g 1000 -S test && adduser -u 1000 -S test -G test
USER test

WORKDIR /home/test

# Setup example directory
COPY --chown=test --from=build /opt/examples/mani.yaml /home/test/

RUN echo 'fpath=( ~/.zsh/completion "${fpath[@]}" ); autoload -Uz compinit && compinit -i' > /home/test/.zshrc
RUN mkdir -p /home/test/.zsh/completion ~/.config/fish/completions
RUN mani completion zsh > /home/test/.zsh/completion/_mani
RUN mani completion fish > ~/.config/fish/completions/mani.fish
RUN echo 'source /etc/profile.d/bash_completion.sh' > /home/test/.bashrc
