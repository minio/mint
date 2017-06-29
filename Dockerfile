FROM ubuntu:16.04

ENV GOROOT /usr/local/go

ENV GOPATH /usr/local

ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH

RUN apt-get update && apt-get install -yq git jq && \
    git clone https://github.com/minio/mint.git && \
    cd /mint && /mint/build/install.sh 

WORKDIR /mint

CMD /mint/run.sh