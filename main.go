package main

import (
	_ "github.com/LevelUp-Journey/Microservice-Analytics/docs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	swagger "github.com/swaggo/fiber-swagger"
)

// @title Microservice Analytics API
// @version 1.0
// @description Analytics microservice using Go Fiber
// @host localhost:3000
// @BasePath /
func main() {
	app := fiber.New()

	app.Use(cors.New())

	app.Get("/swagger/*", swagger.FiberWrapHandler())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Listen(":3000")
}
