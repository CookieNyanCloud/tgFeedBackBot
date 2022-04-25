run:
	go run main.go --local

up:
	docker-compose up --build

.PHONY: run up