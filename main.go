package main

import (
	"context"
	"log"

	"github.com/algo7/day-planner-gpt-data-portal/api/routes"
	redisclient "github.com/algo7/day-planner-gpt-data-portal/internal/redis"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {

	// Check Redis connection.
	err := redisclient.RedisConnectionCheck()
	if err != nil {
		log.Fatalf("Error Connecting to Redis Server: %v", err)
	}

	// Generate an API key as the initial key.
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		log.Fatalf("Error Generating the initial password: %v", err)
	}

	err = redisclient.Rdb.Set(context.Background(), "initial_password", apiKey, 0).Err()
	if err != nil {
		log.Fatalf("Error Setting the initial password in Redis: %v", err)
	}
	log.Printf("Initial Password: %s This will expire once used.", apiKey)

	// Initialize standard Go html template engine
	engine := html.New("./api/assets", ".html")
	engine.Layout("embed") // Optional. Default: "embed"
	engine.Delims("{{", "}}")
	engine.Reload(false)

	// App config.
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Day Planner GPT Data Portal",
		AppName:       "Day Planner GPT Data Portal",
		Views:         engine,
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
