PROJECT_NAME := mtlsproxy

MTLSPROXY_VERSION := 1.2.0
PROJECT_VERSION ?= $(MTLSPROXY_VERSION)-1.0.2

DOCKER_REGISTRY ?= gcr.io/aporetodev
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(PROJECT_VERSION)

build:
	go build

build_linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

package: build_linux
	mkdir -p ./docker/app
	go get -u github.com/agl/extract-nss-root-certs
	curl -s https://hg.mozilla.org/mozilla-central/raw-file/tip/security/nss/lib/ckfw/builtins/certdata.txt -o certdata.txt
	extract-nss-root-certs > docker/app/ca-certificates.pem
	rm -f certdata.txt
	mv ./mtlsproxy ./docker/app

package_fips:
	rm -rf ./docker/fips
	mkdir -p ./docker/fips
	cp -rf main.go internal go.mod go.sum ./docker/fips

container_fips: package_fips
	cd docker && docker build --pull -f Dockerfile.fips -t $(DOCKER_IMAGE_NAME)-fips:$(DOCKER_IMAGE_TAG) .

container: package
	cd docker && docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

$(DOCKER_REGISTRY):
	@docker tag $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) $@/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) \
  && docker push $@/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) \
  && docker tag $(DOCKER_IMAGE_NAME)-fips:$(DOCKER_IMAGE_TAG) $@/$(DOCKER_IMAGE_NAME)-fips:$(DOCKER_IMAGE_TAG) \
  && docker push $@/$(DOCKER_IMAGE_NAME)-fips:$(DOCKER_IMAGE_TAG)

push: container container_fips $(DOCKER_REGISTRY)
