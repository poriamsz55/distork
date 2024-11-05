package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/models/user"
	config "github.com/poriamsz55/distork/configs"
	"github.com/poriamsz55/distork/database"
	"github.com/poriamsz55/distork/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func SignUp(c echo.Context) error {
	username := c.FormValue("username")
	password, err := utils.HashPassword(c.FormValue("password"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Error hashing password",
		})
	}

	email := c.FormValue("email")
	if email == "admin@mail.com" {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Use another email",
		})
	}

	newUser := &user.User{
		Username: username,
		Password: password,
		Email:    email,
		Role:     "USER",
	}

	// Validate user input
	if err := newUser.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Validation failed: " + err.Error(),
		})
	}

	token, err := newUser.GenerateJWT()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Could not generate token",
		})
	}

	collection := database.Collection(config.GetConfigDB().UserColl)

	// check if exists
	var user user.User
	gEmail := fmt.Sprintf("%s@mail.com", c.RealIP())
	err = collection.FindOne(context.TODO(), bson.M{"email": gEmail}).Decode(&user)
	if err == nil {
		// user has guest
		// update the guest
		filter := bson.D{{Key: "email", Value: fmt.Sprintf("%s@mail.com", c.RealIP())}}
		// Creates instructions to add the "avg_rating" field to documents
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "email", Value: newUser.Email},
				{Key: "username", Value: newUser.Username},
				{Key: "password", Value: newUser.Password},
				{Key: "role", Value: newUser.Role},
			}},
		}

		// Updates the first document that has the specified "_id" value
		_, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			// Add user
			err = newUser.AddUserToDB()
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "Unable to create user.",
				})
			}
		}

		// rename upload folder ip@mail.com to user.email...
		err = os.Rename(filepath.Join(config.GetConfigDrive().UploadDir, fmt.Sprintf("%s@mail.com", c.RealIP())),
			filepath.Join(config.GetConfigDrive().UploadDir, newUser.Email))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Couldn't rename directory",
			})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "User created successfully",
			"token":   token,
		})
	}

	// Add user
	err = newUser.AddUserToDB()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Unable to create user.",
		})
	}

	// Create user-specific directory if not exists
	userDir := filepath.Join(config.GetConfigDrive().UploadDir, newUser.Email)
	if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User created successfully",
		"token":   token,
	})
}

func SignIn(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	collection := database.Collection(config.GetConfigDB().UserColl)
	var usr user.User
	err := collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&usr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Invalid username or password",
		})
	}

	if !utils.CheckPasswordHash(password, usr.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"message": "Invalid username or password",
		})
	}

	// Generate JWT token
	token, err := usr.GenerateJWT()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Could not generate token",
		})
	}

	// Send the token in response (no cookie needed)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"token":   token,
	})
}

// Access protected routes and extract user info from JWT
func GetUserProfile(c echo.Context) error {
	usr := c.Get("user").(*user.User)

	return c.JSON(http.StatusOK, usr)
}

func NewGuest(c echo.Context) error {

	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Login successful as guest",
		})
	}

	usr := user.NewUser(c.RealIP(), fmt.Sprintf("%s@mail.com", c.RealIP()), "password", "GUEST")
	err := usr.AddUserToDB()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Add guest user error",
		})
	}

	token, err := usr.GenerateJWT()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Could not generate token",
		})
	}

	// Create upload directory if it doesn't exist
	_, err = os.Stat(filepath.Join(config.GetConfigDrive().UploadDir, fmt.Sprintf("%s@mail.com", c.RealIP())))

	if os.IsNotExist(err) {
		if os.MkdirAll(filepath.Join(config.GetConfigDrive().UploadDir, fmt.Sprintf("%s@mail.com", c.RealIP())), os.ModePerm) != nil {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"message": "Coludn't create directory",
			})
		}
	}

	// Send the token in response (no cookie needed)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"token":   token,
	})
}
