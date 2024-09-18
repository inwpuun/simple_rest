package main

import (
	"context"
	"fmt"
	db "inwpuun/simplerest/db/generate"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
)

type UserRequest struct {
	Name string `json:"name"`
	Bio  string `json:"bio"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// fiber instance
	app := fiber.New()

	// urlExample := "postgres://username:password@localhost:5432/database_name"
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	q := db.New(conn)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/users", func(c *fiber.Ctx) error {
		users, err := q.ListUsers(context.Background())
		if err != nil {
			c.Status(fiber.StatusInternalServerError).SendString("Failed to list users")
		}
		return c.JSON(users)
	})

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid user ID",
			})
		}

		user, err := q.GetUser(context.Background(), int64(id))
		switch err {
		case nil:
		case pgx.ErrNoRows:
			return c.Status(fiber.StatusNotFound).SendString("User not found")
		default:
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to get user")
		}

		return c.JSON(user)
	})

	app.Post("/users", func(c *fiber.Ctx) error {
		body := new(UserRequest)
		if err := c.BodyParser(body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}

		if body.Name == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Name is required")
		}

		err := q.CreateUser(context.Background(), db.CreateUserParams{Name: body.Name, Bio: pgtype.Text{String: body.Bio, Valid: true}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to create user")
		}
		return c.Status(fiber.StatusCreated).SendString("User created")
	})

	app.Put("/users/:id", func(c *fiber.Ctx) error {
		// Parse user ID from the URL
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid user ID",
			})
		}

		// Parse the request body into a UserUpdateRequest struct
		var req UserRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}

		_, err = q.GetUser(context.Background(), int64(id))
		switch err {
		case nil:
			// Execute the SQLC-generated UpdateUser method
			err = q.UpdateUser(c.Context(), db.UpdateUserParams{
				ID:   int64(id),
				Name: req.Name,
				Bio:  pgtype.Text{String: req.Bio, Valid: req.Bio != ""},
			})
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to update user",
				})
			}

			// Return a success response
			return c.Status(fiber.StatusNoContent).SendString("User updated")
		case pgx.ErrNoRows:
			err := q.CreateUser(context.Background(), db.CreateUserParams{Name: req.Name, Bio: pgtype.Text{String: req.Bio, Valid: true}})
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Failed to create user")
			}
			return c.Status(fiber.StatusCreated).SendString("User created")
		default:
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to get user")
		}
	})

	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid user ID")
		}

		err = q.DeleteUser(context.Background(), id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete user")
		}
		return c.Status(fiber.StatusNoContent).SendString("User deleted")
	})

	log.Fatal(app.Listen(
		fmt.Sprintf(":%s", os.Getenv("PORT")),
	))
}
