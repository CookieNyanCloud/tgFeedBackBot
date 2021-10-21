run:
	go run main.go

build:
	docker-compose up --build

.PHONY: run build