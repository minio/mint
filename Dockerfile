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
  build-essential \
  python3 \
  python3-pip \
  golang-go && \
  apt-get install -y \
  jq && \ 
  #libssl-dev && \
  openssl && \

  go get -u github.com/minio/minio-go && \
  go get -u github.com/minio/minio-go && \
  chmod +x run.sh 

 RUN  pip3 install -r /tmp/requirements.txt
 RUN  pip3 install yq
 RUN  pip3 install minio
 RUN  rm /tmp/requirements.txt 


CMD ./run.sh
   