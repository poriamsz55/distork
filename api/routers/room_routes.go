package router

import (
	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/handlers"
	"github.com/poriamsz55/distork/api/models/room"
)

func WSRoutes(e *echo.Group) {
	//
	e.GET("", handlers.HandeConnection)
	e.GET("/room/create", room.CreateRoom)
}
