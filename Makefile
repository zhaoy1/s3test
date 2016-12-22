all: deps install

deps:
	go get -v github.com/aws/aws-sdk-go/aws
	go get -v github.com/aws/aws-sdk-go/aws/credentials
	go get -v github.com/aws/aws-sdk-go/aws/session
	go get -v github.com/aws/aws-sdk-go/service/s3
	go get -v github.com/ghodss/yaml
	go get -v github.com/influxdata/influxdb/client/v2

fmt:
	@ if [ ! $$(gofmt -e -d . | wc -l) -eq 0 ]; then \
		>&2 echo "gofmt failed (reproduce \`gofmt -e -d *.go\`):" ;\
		>&2 gofmt -e -d *.go ; exit 1 ; \
	fi

install: deps fmt
	go build 
