auth-up:
	docker-compose -f ./src/auth/docker-compose.yml up -d
	migrate -source "file://./src/auth/storage/migrations" -database "postgres://postgres:PGPWD123@localhost:7001/postgres?sslmode=disable" up

auth-down:
	docker-compose -f ./src/auth/docker-compose.yml kill
