package routes

import (
	"github.com/algo7/day-planner-gpt-data-portal/api/controllers"
	"github.com/gofiber/fiber/v2"
)

// EmailsRoutes is the route handler for the emails API.
func EmailsRoutes(app *fiber.App) {
	app.Get("/outlook", controllers.GetOutlookEmails)
}
