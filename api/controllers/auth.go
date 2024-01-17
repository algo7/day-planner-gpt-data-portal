package controllers

import "github.com/gofiber/fiber/v2"

func GetAuth(c *fiber.Ctx) error {
	return c.SendString("Auth")
}


