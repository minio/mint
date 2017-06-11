FROM ubuntu:17.04
COPY . /home
COPY ./Apps/minio-py/requirements.txt /tmp/requirements.txt
WORKDIR /home
ENV GOPATH=/home
ENV PATH=$PATH:$GOPATH/bin
#apt-get upgrade -y 

RUN apt-get update && apt-get install -y \
  git \
  bash \
  curl \
  sudo \
  build-essential \
  python3 \
  python3-pip \
  golang-go && \
  apt-get install -y \
  jq && \ 
  openssl && \

  pip3 install yq && \
  curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash - && \
  sudo apt-get install -y nodejs && \
  sudo apt-get install -y  maven && \
  sudo apt-get install -y default-jre default-jdk && \
  apt-get clean && \ 
  update-alternatives --config java && update-alternatives --config javac && \
  chmod +x ./run.sh

CMD ./run.sh
   