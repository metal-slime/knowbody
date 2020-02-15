#!make
include .env
export $(shell sed 's/=.*//' .env)

.PHONY: check run

check: 
	golangci-lint run -c .golang-ci.yml ./... 
	go run cmd/main.go --lint

test:
	go test -v ./...

run:
	go run cmd/main.go

build-image:
	docker build -t knowbody .