package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/monster-monitoring-service/handlers"
	"github.com/byuoitav/monster-monitoring-service/salt"
	"github.com/byuoitav/monster-monitoring-service/store"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {

	store.OnStart()

	var control sync.WaitGroup
	NUM_PROCESSES := 2

	signals := make(chan os.Signal, 1)
	timer := make(chan bool, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		log.Printf("Waiting for SIGTERM signal...")
		output := <-signals
		log.Printf("output: %v", output)
		log.Printf("Nuclear launch detected. Firing interceptors...")
		for i := 0; i < NUM_PROCESSES; i++ {
			timer <- true
		}
		control.Wait()
		os.Exit(0)
	}()

	events := make(chan salt.SaltEvent)
	control.Add(NUM_PROCESSES)
	go salt.Listen(events, timer, &control)
	go store.Listen(events, timer, &control)

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
