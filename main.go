package main

import "github.com/gofiber/fiber/v2"

func main() {

	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Day Planner GPT Data Portal",
		AppName:       "Day Planner GPT Data Portal",
	})

	app.Listen(":3000")

}
