package routes

import (
	"github.com/algo7/day-planner-gpt-data-portal/api/controllers"
	"github.com/gofiber/fiber/v2"
)

// HomeRoutes is the route handler for the home page.
func HomeRoutes(app *fiber.App) {
	app.Get("/", controllers.GetHome)
}
