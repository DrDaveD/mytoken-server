package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/zachmann/mytoken/internal/endpoints/configuration"
	"github.com/zachmann/mytoken/internal/endpoints/redirect"
)

var server *fiber.App

func Init() {
	server = fiber.New(fiber.Config{
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   90 * time.Second,
		IdleTimeout:    150 * time.Second,
		ReadBufferSize: 8192,
		//WriteBufferSize: 4096,
	})
	addMiddlewares(server)
	addRoutes(server)
}

func addRoutes(s fiber.Router) {
	s.Get("/", handleTest)
	s.Get("/test", handleTest)
	s.Get("/.well-known/mytoken-configuration", configuration.HandleConfiguration)
	s.Get("/.well-known/openid-configuration", func(ctx *fiber.Ctx) error {
		return ctx.Redirect("/.well-known/mytoken-configuration")
	})
	s.Get("/redirect", redirect.HandleOIDCRedirect)
	addAPIRoutes(s)
}

func start(s *fiber.App) {
	log.Fatal(s.Listen(":8000"))
}

func Start() {
	start(server)
}
