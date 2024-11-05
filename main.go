package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
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

	usr := user.NewUser("admin", "admin@mail.com", "admin", "ADMIN")
	err = usr.AddUserToDB()
	if err != nil {
		return
	}

	e := echo.New()
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
	// middle.JWTMiddleWares(userGroup)
	router.UserRoutes(userGroup)

	// Start server
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  2 * time.Hour, // Adjust as necessary
		WriteTimeout: 2 * time.Hour, // Adjust as necessary
	}
	e.Logger.Fatal(e.StartServer(server))
}
