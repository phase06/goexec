package main

import (
	"goexec/execServer/handler"
	"goexec/execServer/helper"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	badger "github.com/dgraph-io/badger/v2"
)

// Version of current package
const Version = "0.5.0"

func helloWorld(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")

}

func setupRoutes(app *fiber.App) {
	app.Get("/", helloWorld)

	app.Get("/api/result/:id", handler.CheckStatus)
	app.Post("/api/execute", handler.QueueTask)
}

func main() {

	helper.StartupMessage(":3000", Version)
	// open db
	opt := badger.DefaultOptions("db").WithTruncate(true).WithValueLogFileSize(256 << 20)
	var err error
	handler.TaskDB, err = badger.Open(opt)
	if err != nil {
		log.Fatal(err)
	}

	defer handler.TaskDB.Close()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(logger.New())

	setupRoutes(app)
	app.Listen(":3000")

}
