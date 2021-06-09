FROM alpine:3.10 as build

COPY --from=golang:1.16.3-alpine /usr/local/go/ /usr/local/go/

ENV XDG_CACHE_HOME=/tmp/.cache
ENV GOPATH=${HOME}/go
ENV GO111MODULE=on
ENV PATH="/usr/local/go/bin:${PATH}"

RUN apk update
RUN apk add --no-cache make build-base bash curl g++ git

WORKDIR /opt

COPY . .

RUN make build

FROM alpine:3.10

RUN apk update
RUN apk add --no-cache sudo bash zsh fish bash-completion git

# Copy executable
COPY --from=build /opt/execs/mani /usr/local/bin/mani

RUN mani completion bash > /usr/share/bash-completion/completions/mani

RUN addgroup -g 1000 -S test && adduser -u 1000 -S test -G test
USER test

WORKDIR /home/test

COPY --chown=test --from=build /opt/_example/mani.yaml /opt/_example/.gitignore /home/test/
COPY --chown=test --from=build /opt/test/.zshrc /home/test/.zshrc

RUN mkdir -p /home/test/.zsh/completion ~/.config/fish/completions
RUN mani completion zsh > /home/test/.zsh/completion/_mani
RUN mani completion fish > ~/.config/fish/completions/mani.fish
RUN echo 'source /etc/profile.d/bash_completion.sh' > /home/test/.bashrc