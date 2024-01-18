package main

import (
	"log"

	"github.com/algo7/day-planner-gpt-data-portal/api/routes"
	redisclient "github.com/algo7/day-planner-gpt-data-portal/internal/redis"
	"github.com/gofiber/fiber/v2"
)

func main() {

	// Check Redis connection.
	err := redisclient.RedisConnectionCheck()
	if err != nil {
		log.Fatalf("Error Connecting to Redis Server: %v", err)
	}

	// App config.
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Day Planner GPT Data Portal",
		AppName:       "Day Planner GPT Data Portal",
	})

	// Load the routes.
	routes.HomeRoutes(app)
	routes.EmailsRoutes(app)
	routes.AuthRoutes(app)

	// Start the server.
	err = app.Listen(":3000")

	if err != nil {
		log.Fatalf("Error Starting the Server: %v", err)
	}

}
