package main

import (
	"errors"
	"fmt"
	"log"
	"time"
	"url-shortener/database"
	"url-shortener/handlers"
	"url-shortener/helpers"
	"url-shortener/logs"
	"url-shortener/routes"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	app     *fiber.App
	factory *handlers.Factory
)

func bootstrap() {
	logs.NewLogger()
	defer logs.Close()

	initTimezone()

	db := database.NewConnection()
	rdb := database.NewRedis()
	factory = handlers.NewFactory(db, rdb)

	initRoutes()
}

func initRoutes() {
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(helmet.New())

	api := app.Group("/api")
	routes.NewWebRoutes(app, factory)
	routes.NewAPIRoutes(api, factory)
}

func initTimezone() {
	tz := helpers.Getenv("APP_TIMEZONE", "")
	loc, err := time.LoadLocation(tz)
	if err != nil {
		panic(fmt.Sprintf("can't set timezone to %v: %v", tz, err))
	}
	time.Local = loc
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"message": message,
	})
}

func main() {
	app = fiber.New(fiber.Config{
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
		ErrorHandler: errorHandler,
	})

	bootstrap()

	port := helpers.Getenv("APP_PORT", "5000")
	err := app.Listen(":" + port)
	if err != nil {
		log.Fatalf("failed to listen on port %v: %v", port, err)
	}
}
