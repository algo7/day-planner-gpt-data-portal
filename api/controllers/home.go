package controllers

import (
	"github.com/gofiber/fiber/v2"
)

// GetHome returns the home page.
// @Summary Get Home
// @ID getHome
// @Description This endpoint returns a welcome message.
// @Tags Home
// @Accept json
// @Produce json
// @Success 200 {object} Response "Returns a welcome message"
// @Router / [get]
func GetHome(c *fiber.Ctx) error {

	return c.Status(fiber.StatusOK).JSON(Response{Data: "Welcome to the Fiber API"})
}
