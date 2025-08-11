.PHONY: build
build:
	docker-compose build

.PHONY: up
up:
	docker-compose up --build

.PHONY: test
test:
	go test ./... || true
