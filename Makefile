migrate-pg:
	migrate -source "file://./auth/storage/migrations" -database "postgres://postgres:PGPWD123@localhost:7001/postgres?sslmode=disable" up
