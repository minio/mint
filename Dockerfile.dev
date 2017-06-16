FROM ubuntu:16.04
ENV GOROOT /usr/local/go
ENV GOPATH /usr/local
ENV PATH PATH=$GOPATH/bin:$GOROOT/bin:$PATH
RUN apt-get update && apt-get install -yq \
  curl \
  default-jre \ 
  default-jdk \
  git \
  jq \
  python3 \
  openssl \
  sudo \
  python3-pip && \
  pip3 install yq && \
  curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash - && \
  sudo apt-get install -y nodejs && \ 
  curl -O https://storage.googleapis.com/golang/go1.7.4.linux-amd64.tar.gz && \
  tar -xf go1.7.4.linux-amd64.tar.gz && \
  mv go /usr/local && \
  rm -rf go && \
  rm go1.7.4.linux-amd64.tar.gz
COPY . /mint
WORKDIR /mint
RUN chmod +x ./run.sh
CMD ./run.sh