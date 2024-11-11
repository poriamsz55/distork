package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	middle "github.com/poriamsz55/distork/api/middlewares"
	"github.com/poriamsz55/distork/api/models/distork"
	"github.com/poriamsz55/distork/api/models/user"
	router "github.com/poriamsz55/distork/api/routers"
	config "github.com/poriamsz55/distork/configs"
	"github.com/poriamsz55/distork/database"
)

func main() {
	// Create upload directory if it doesn't exist
	if _, err := os.Stat(config.GetConfigDrive().UploadDir); os.IsNotExist(err) {
		os.MkdirAll(config.GetConfigDrive().UploadDir, os.ModePerm)
	}

	_, err := database.Connect()
	if err != nil {
		log.Fatalf("Error when opening file: %s", err)
		return
	}

	usr := user.NewUser("admin",
		"admin@mail.com",
		"admin",
		config.RoleAdmin)

	err = usr.AddUserToDB()
	if err != nil {
		return
	}

	e := echo.New()

	setupApp(e)

	// e.Logger.Fatal(e.StartTLS(":8080",
	// 	"localhost+2.pem",
	// 	"localhost+2-key.pem"))

	e.Logger.Fatal(e.Start(":8080"))
}

func setupApp(e *echo.Echo) {
	// Use ExtractIPFromXFFHeader to properly handle IPs behind a proxy
	e.IPExtractor = echo.ExtractIPFromXFFHeader()

	// main routes
	// api to handle server side (nginx)
	eGroup := e.Group("/api")
	middle.MiddleWares(eGroup)
	router.Routes(eGroup)

	// drive upload route
	// Applying the rate limiting middleware only to the upload route
	uploadGroup := eGroup.Group("/drive/upload")
	middle.JWTMiddleWares(uploadGroup)
	middle.UploadMiddleWares(uploadGroup)
	router.UploadRoutes(uploadGroup)

	// Drive Routes
	driveGroup := eGroup.Group("/drive")
	middle.JWTMiddleWares(driveGroup)
	router.DriveRoutes(driveGroup)

	// User Routes
	userGroup := eGroup.Group("/user")
	router.UserRoutes(userGroup)

	// Room
	//create new manager for websocket traffic
	hub := distork.NewHub()
	go hub.Run()
	websokcetGroup := eGroup.Group("/ws")
	middle.WSJWTMiddleWares(websokcetGroup)
	router.DistorkRoutes(websokcetGroup, hub)

}
