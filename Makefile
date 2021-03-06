# ===================================
# ======= Local Configuration =======
# ===================================

VERSION := $(shell git tag | tail -n1)
BINARY_NAME = beanpay-api
BINARY_VERSIONED = ${BINARY_NAME}-${VERSION}

start: clean build run

clean:
	rm -f ${BINARY_NAME}-*

build:
	go build -o ${BINARY_VERSIONED}

run:
	./${BINARY_VERSIONED}

test:
	@go test ./... -race -coverprofile=coverage.txt -covermode=atomic
	go tool cover -html=coverage.txt
	rm coverage.txt

.PHONY: start clean build run test

# ====================================
# ======= Docker Configuration =======
# ====================================

docker: docker-build
	docker-compose up

docker-build:
	docker-compose build

docker-clean:
	docker-compose down

docker-publish:
	docker tag beanpay/api gcr.io/beanpay/api
	docker push gcr.io/beanpay/api

.PHONY: docker docker-build docker-clean
