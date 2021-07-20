-- migrate create -ext sql -dir="./src/auth/storage/migrations" create_users_table

CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    email VARCHAR (50) UNIQUE NOT NULL,
    username VARCHAR (50) UNIQUE NOT NULL,
    password VARCHAR (128) UNIQUE NOT NULL
);
