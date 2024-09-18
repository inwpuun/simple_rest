start:
	go run main.go

migrate-up:
	migrate -path db/migrations -database "postgresql://root:password@localhost:5432/gorm?sslmode=disable" up

migrate-down:
	migrate -path db/migrations -database "postgresql://root:password@localhost:5432/gorm?sslmode=disable" down

generate:
	sqlc generate
