package middlewares

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/models/user"
	config "github.com/poriamsz55/distork/configs"
)

func JWTMiddleWares(e *echo.Group) {
	// JWT
	e.Use(OptionalJWTMiddleware)
}

func OptionalJWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")

		usr := user.NewUser(c.RealIP(),
			fmt.Sprintf("%s@mail.com", c.RealIP()),
			"password",
			config.RoleGuest)

		var err error
		if authHeader != "" {
			// Try to verify JWT
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			err = verifyToken(tokenString)
			if err == nil {
				usr, err = user.GetUserByToken(tokenString)
				if err != nil {
					if usr.AddUserToDB() != nil {
						return err
					}

				}

			} else {
				if usr.AddUserToDB() != nil {
					return err
				}
			}
		} else {
			if usr.AddUserToDB() != nil {
				return err
			}
		}

		c.Set("user", usr)
		// If no token or invalid token, the request proceeds as a guest
		return next(c)
	}
}

func verifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return config.GetSharedConfig().JwtSecret, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}
