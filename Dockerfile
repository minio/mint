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
  go get -u github.com/minio/minio-go && \
  go get -u github.com/minio/minio-go && \
  chmod +x run.sh 

 RUN curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash -
 RUN sudo apt-get install -y nodejs 

 RUN  pip3 install -r /tmp/requirements.txt
 RUN  pip3 install yq
 RUN  pip3 install minio
 RUN  rm /tmp/requirements.txt 

 RUN sudo apt-get install -y  maven

ENV MAVEN_HOME /opt/maven

RUN sudo apt-get install -y default-jre default-jdk
# remove download archive files
RUN apt-get clean


# configure symbolic links for the java and javac executables
RUN update-alternatives --config java && update-alternatives --config javac

CMD ./run.sh
   