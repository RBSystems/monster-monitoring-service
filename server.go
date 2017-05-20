package main

import (
	"net/http"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/monster-monitoring-service/badger"
	"github.com/byuoitav/monster-monitoring-service/handlers"
	"github.com/byuoitav/monster-monitoring-service/helpers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {

	//initialize the badger store
	badger.Init()

	//get the status of every room and building from Configuration-Database
	helpers.OnStart()

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
