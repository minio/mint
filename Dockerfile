
FROM ubuntu:16.04

ENV GOPATH /go
ENV PATH $PATH:$GOPATH/bin

RUN apt-get update && apt-get install -yq \
  curl \
  default-jre \ 
  default-jdk \
  git \
  golang-go \
  jq \
  nodejs \ 
  npm \
  python3 \
  openssl \
  python3-pip && \
  pip3 install yq && \
  update-alternatives --config java && update-alternatives --config javac


COPY . /mint
WORKDIR /mint

RUN chmod +x ./run.sh

CMD ./run.sh
   