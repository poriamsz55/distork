package router

import (
	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/handlers"
	middle "github.com/poriamsz55/distork/api/middlewares"
)

func UserRoutes(e *echo.Group) {
	// Protect this route, accessible only with a valid token
	e.GET("/profile", handlers.GetUserProfile, middle.OptionalJWTMiddleware)

	// Protect this route, accessible only with a valid token
	e.GET("/profile/update", handlers.UpdateProfile, middle.OptionalJWTMiddleware)

	// Sign up
	e.POST("/signup", handlers.SignUp)

	// Sign in
	e.POST("/signin", handlers.SignIn)
}
