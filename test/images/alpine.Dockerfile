FROM alpine:3.10

COPY --from=golang:1.16.3-alpine /usr/local/go/ /usr/local/go/

ENV GOPATH=${HOME}/go
ENV GO111MODULE=on
ENV PATH="/usr/local/go/bin:${PATH}"

RUN apk update
RUN apk add make bash g++

# Mock git
RUN echo -e '#!/bin/bash\ngit() { echo 123; }' > /usr/bin/git && chmod +x /usr/bin/git

WORKDIR /opt

COPY . .

RUN go mod download

# CMD ["/bin/bash"]
# CMD ["/bin/bash", "git"]
# CMD ["go", "test", "-v", "-cover", "./..."]
