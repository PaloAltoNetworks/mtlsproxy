MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash -o pipefail

build:
	go build

build_linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

package: build_linux
	rm -rf ./docker/app
	mkdir -p ./docker/app
	go get -u github.com/agl/extract-nss-root-certs
	curl -s https://hg.mozilla.org/mozilla-central/raw-file/tip/security/nss/lib/ckfw/builtins/certdata.txt -o certdata.txt
	extract-nss-root-certs > docker/app/ca-certificates.pem
	rm -f certdata.txt
	mv ./mtlsproxy ./docker/app

container: package
	cd docker && docker build -t gcr.io/aporetodev/mtlsproxy:1.1.0 .
