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
	"github.com/redis/go-redis/v9"
)

// @title Fiber Example API
// @version 1.0
// @description day-planner-gpt-data-portal
// @termsOfService http://swagger.io/terms/
// @contact.name Algo7
// @contact.email tools@algo7.tools
// @license.name The GNU General Public License v3.0
// @license.url https://raw.githubusercontent.com/algo7/day-planner-gpt-data-portal/main/LICENSE
// @host localhost:3000
// @BasePath /
func main() {

	// Check if google_credentials.json and outlook_credentials.json exist in the ./credentials folder.
	gExists := utils.FileExists("./credentials/google_credentials.json")
	if !gExists {
		log.Fatal("Error: google_credentials.json not found in ./credentials folder.")
	}

	oExists := utils.FileExists("./credentials/outlook_credentials.json")
	if !oExists {
		log.Fatal("Error: outlook_credentials.json not found in ./credentials folder.")
	}

	// Check Redis connection.
	err := redisclient.RedisConnectionCheck()
	if err != nil {
		log.Fatalf("Error Connecting to Redis Server: %v", err)
	}

	// Check if the initial password is already set in Redis.
	initialPassword, err := redisclient.Rdb.Get(context.Background(), "initial_password").Result()
	if err != nil {
		// If the initial password is not set in Redis, generate one and set it.
		if err == redis.Nil {
			log.Println("Initial Password not set in Redis. Generating...")

			// Generate an API key as the initial password.
			apiKey, err := utils.GenerateAPIKey()
			if err != nil {
				log.Fatalf("Error Generating the initial password: %v", err)
			}

			// Set the initial password in Redis.
			err = redisclient.Rdb.Set(context.Background(), "initial_password", apiKey, 0).Err()
			if err != nil {
				log.Fatalf("Error Setting the initial password in Redis: %v", err)
			}
			log.Printf("Initial Password: %s This will expire once used.", apiKey)
		}

		log.Fatalf("Error checking the initial password in Redis: %v", err)
	}

	// If the initial password is not an empty string, it means that it has been used.
	if initialPassword == "" {
		log.Print("Initial password has been used. No need to generate a new one.")
	}

	// Initialize standard Go html template engine
	engine := html.New("./assets", ".html")
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

	// Swagger middleware
	app.Use(swagger.New(swagger.ConfigDefault))

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
