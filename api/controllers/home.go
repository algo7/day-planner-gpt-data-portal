package controllers

import (
	"github.com/gofiber/fiber/v2"
)

// GetHome returns the home page.
// @Summary Home Page
// @Description Returns the home page of the application.
// @Tags Pages
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]string "data: 'Hello, World!'"
// @Router / [get]
func GetHome(c *fiber.Ctx) error {
	return c.JSON(map[string]string{"data": "Hello, World!"})
}
