package router

import (
	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/handlers"
)

func WSRoutes(e *echo.Group) {
	//
	e.GET("", handlers.HandleWS)
}
