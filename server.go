package main

import (
	"net/http"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/monster-monitoring-service/handlers"
	"github.com/byuoitav/monster-monitoring-service/helpers"
	"github.com/byuoitav/monster-monitoring-service/salt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {

	helpers.OnStart()

	timer := make(chan bool, 1)

	go salt.Start(timer)

	go salt.Listen()

	port := ":10000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	// Use the `secure` routing group to require authentication
	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	secure.GET("/buildings/:building/rooms/:room", handlers.ViewRoom)

	secure.Static("/", "dist")

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}
