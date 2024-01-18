package controllers

import (
	"github.com/gofiber/fiber/v2"
)

// GetHome returns the home page.
// @Summary Home Page
// @Description Returns the home page of the application.
// @Tags Pages
// @Accept */*
// @Produce plain
// @Success 200 {string} string "Hello, World!"
// @Router / [get]
func GetHome(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}
