FROM ubuntu:16.04

ENV DEBIAN_FRONTEND noninteractive

ENV LANG C.UTF-8

ENV GOROOT /usr/local/go

ENV GOPATH /usr/local

ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH

RUN apt-get --yes update && apt-get --yes upgrade && apt-get --yes --quiet install wget jq curl git && \
    git clone https://github.com/minio/mint.git /mint && \
    cd /mint && /mint/release.sh

WORKDIR /mint

ENTRYPOINT ["/mint/entrypoint.sh"]
