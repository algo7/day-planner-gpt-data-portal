package main

import (
	"log"
	"os"

	"github.com/algo7/day-planner-gpt-data-portal/api/routes"
	"github.com/algo7/day-planner-gpt-data-portal/pkg/integrations/outlook"
	"github.com/gofiber/fiber/v2"
)

func main() {

	// fmt.Println("Enter Oauth Token: ")
	// var input string
	// _, err := fmt.Scan(&input)
	// if err != nil {
	// 	log.Fatalf("Error Scanning Input: %v", err)
	// }
	token := os.Getenv("OAUTH_TOKEN")
	err := outlook.GetEmails(token)
	if err != nil {
		log.Fatalf("Error Getting Emails: %v", err)
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

	// Start the server.
	err = app.Listen(":3000")

	if err != nil {
		log.Fatalf("Error Starting the Server: %v", err)
	}

}
