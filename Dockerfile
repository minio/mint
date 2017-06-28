FROM ubuntu:16.04

ENV GOROOT /usr/local/go

ENV GOPATH /usr/local

ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH

WORKDIR /mint

RUN apt-get update && apt-get install -yq \
    git && git clone https://github.com/minio/mint.git && \
    cd /mint && /mint/build/install.sh 

CMD /mint/run.sh