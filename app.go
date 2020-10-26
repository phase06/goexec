package main

import (
	"flag"
	"goexec/execServer/handler"
	"goexec/execServer/helper"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	badger "github.com/dgraph-io/badger/v2"
)

var mutex = &sync.Mutex{}

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

func startTaskPolling() {
	for {
		time.Sleep(5 * time.Second)
		// to ensure only one task is being executed
		mutex.Lock()
		handler.ProcessQueue()
		mutex.Unlock()
	}
}

func main() {

	// cmdline arg
	var portFlag = flag.String("port", "3000", "Port to start this app")
	var dropFlag = flag.Bool("drop", false, "true to drop all datas")
	flag.Parse()

	defaultPort := "3000"
	if *portFlag != "" {
		defaultPort = *portFlag
	}
	helper.StartupMessage(":"+defaultPort, Version)
	// open db
	opt := badger.DefaultOptions("db").WithTruncate(true).WithValueLogFileSize(256 << 20)
	var err error
	handler.TaskDB, err = badger.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
	if *dropFlag {
		handler.TaskDB.DropAll()
	}

	defer handler.TaskDB.Close()

	go startTaskPolling()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(logger.New())

	setupRoutes(app)
	app.Listen(":3000")

}
