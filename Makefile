.PHONY: db-up db-down infra-up infra-down

db-up:
	docker run --name envman-db -e POSTGRES_USER=envman -e POSTGRES_PASSWORD=envman123 -e POSTGRES_DB=envman -p 5432:5432 -d postgres:16

db-down:
	docker stop envman-db
	docker rm envman-db

infra-up: db-up

infra-down: db-down