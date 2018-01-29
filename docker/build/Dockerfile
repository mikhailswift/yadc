FROM centos:7
MAINTAINER Mikhail Swift <mikhail.swift@gmail.com>

WORKDIR /tmp/

ARG GO_VER=1.9.3

ENV PATH=/usr/local/go/bin/:$PATH
ENV GOPATH=/go
ENV GOBIN=/go/bin
ENV PATH=$GOBIN:$PATH

RUN mkdir -p $GOPATH $GOBIN

ADD https://dl.google.com/go/go${GO_VER}.linux-amd64.tar.gz go.tar.gz
RUN tar -C /usr/local -xzf go.tar.gz && rm go.tar.gz

RUN yum -y \
    make \
    git

RUN go get -u github.com/golang/lint/golint
