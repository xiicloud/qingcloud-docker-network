FROM golang:1.7-alpine
MAINTAINER Shijiang Wei <mountkin@gmail.com>

ARG GIT_COMMIT

ADD . $GOPATH/src/github.com/nicescale/qingcloud-docker-network
WORKDIR $GOPATH/src/github.com/nicescale/qingcloud-docker-network
RUN CGO_ENABLED=0 go build \
  -ldflags "-w -s -X main.version=$(cat VERSION) -X main.gitCommit=$GIT_COMMIT" \
  -o /bin/qingcloud-docker-network
