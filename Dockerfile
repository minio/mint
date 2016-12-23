FROM golang:1.7-alpine

ADD . /home

WORKDIR /home/minio-go-functional-test/

RUN \
       apk add --no-cache bash git openssh && \
       go get -u github.com/minio/minio-go && \
       go test -c api_functional_v4_test.go 

CMD ["./minio.test", "-test.timeout", "3600s","-test.v","-test.run","Test*"]
