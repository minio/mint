FROM golang:1.7-alpine

COPY . /home

WORKDIR /home

# Install commonly used dependencies here.
# Any unique depencies can be installed in respective build.sh files 

RUN \
       apk add --no-cache bash git openssh mailcap && \
       go get -u github.com/minio/minio-go && \
       chmod +x run.sh

CMD ./run.sh
   
    
