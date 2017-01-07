FROM golang:1.7-alpine

COPY . /home

WORKDIR /home

RUN \
       apk add --no-cache bash git openssh mailcap && \
       go get -u github.com/minio/minio-go && \
       go test -c /home/minio-functional-test/server_test.go && \
       go test -c /home/minio-go-functional-test/api_functional_v4_test.go 

CMD /home/cmd.test -test.v && \ 
    /home/minio.test -test.timeout 3600s -test.v 
    
