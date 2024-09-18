run docker
`docker run --name simple-database -e POSTGRES_USER=root -e POSTGRES_PASSWORD=password -e POSTGRES_DB=gorm -d postgres`

install dependency
`go mod tidy`

generate sqlc
`make generate`

migrate database
`make migrate-up`

run server
`make start`