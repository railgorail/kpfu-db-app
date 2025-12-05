db:
	docker-compose up --build
tidy:
	go mod tidy
run:
	go run cmd/main.go
compose-up:
	docker-compose up --build
compose-down:
	docker-compose down
.PHONY: db tidy run compose-up compose-down