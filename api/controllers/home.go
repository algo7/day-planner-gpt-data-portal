package controllers

import "github.com/gofiber/fiber/v2"

// GetHome returns the home page.
func GetHome(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}
