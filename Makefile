all: deps install

deps:
	go get -v github.com/aws/aws-sdk-go/aws
	go get -v github.com/aws/aws-sdk-go/aws/credentials
	go get -v github.com/aws/aws-sdk-go/aws/session
	go get -v github.com/aws/aws-sdk-go/service/s3

fmt:
	@ if [! $$(gofmt -e -d *.go |wc -l) -eq 0]; then \
		>&2 echo "gofmt failed (reproduce \`gofmt -e -d *.go\`):" ;\
		>&2 gofmt -e -d *.go ; exit 1 ; \

install:
	go build 
