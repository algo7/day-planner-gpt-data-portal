package main

import (
	"context"
	"log"

	"github.com/algo7/day-planner-gpt-data-portal/api/middlewares"
	"github.com/algo7/day-planner-gpt-data-portal/api/routes"
	redisclient "github.com/algo7/day-planner-gpt-data-portal/internal/redis"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/utils"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/gofiber/template/html/v2"
)

// @title Fiber Example API
// @version 1.0
// @description day-planner-gpt-data-portal
// @termsOfService http://swagger.io/terms/
// @contact.name Algo7
// @contact.email tools@algo7.tools
// @license.name GNU GENERAL PUBLIC LICENSE Version 3
// @license.url https://raw.githubusercontent.com/algo7/day-planner-gpt-data-portal/main/LICENSE
// @host localhost:3000
// @BasePath /
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

	// Auth middleware
	app.Use(keyauth.New(keyauth.Config{
		Next:      middlewares.AuthFilter,
		KeyLookup: "header:X-API-KEY",
		Validator: middlewares.ValidateAPIKey,
	}))

	// Healthcheck middleware /livez and /readyz routes
	app.Use(healthcheck.New())

	app.Use(swagger.New(swagger.Config{
		FilePath: "./docs/swagger.json",
	}))

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
