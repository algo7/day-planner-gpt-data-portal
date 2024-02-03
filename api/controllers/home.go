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
// @Success 200 {object} APIResponse
// @Router / [get]
func GetHome(c *fiber.Ctx) error {
	response := APIResponse{
		Status: 200,
		Data:   map[string]string{"data": "Hello, World!"},
	}
	return c.JSON(response)
}
