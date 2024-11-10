package router

import (
	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/handlers"
	"github.com/poriamsz55/distork/api/models/distork"
)

func DistorkRoutes(e *echo.Group, hub *distork.Hub) {
	//
	e.GET("", func(c echo.Context) error {
		return handlers.HandleConnection(c, hub)
	})
}
