build:
	docker compose build
start: build
	docker compose up
fmt:
	golangci-lint run
fmt-fix:
	golangci-lint run --fix