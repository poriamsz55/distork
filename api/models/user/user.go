package user

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt"
	config "github.com/poriamsz55/distork/configs"
	"github.com/poriamsz55/distork/database"
	"github.com/poriamsz55/distork/utils"
	"go.mongodb.org/mongo-driver/bson"
)

type User struct {
	Username  string `json:"username" bson:"username" validate:"required,min=3,max=30"`
	Email     string `json:"email" bson:"email" validate:"required,email"`
	Password  string `json:"password,omitempty" bson:"password" validate:"required,min=8"`
	Role      string `json:"role" bson:"role"`
	DriveSize int64  `json:"drive_size,omitempty" bson:"drive_size"`
	DriveUsed int64  `json:"drive_used,omitempty" bson:"drive_used"`
}

func UpdateUser(username string, updateFields bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"username": username}
	update := bson.M{"$set": updateFields}

	collection := database.Collection(config.GetConfigDB().UserColl)
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func NewUser(username, email, password, role string) *User {
	pass, _ := utils.HashPassword(password)
	usr := &User{
		Username:  username,
		Email:     email,
		Password:  pass,
		Role:      role,
		DriveSize: config.RoleDriveSize[role],
		DriveUsed: 0,
	}

	return usr
}

func (u *User) GenerateJWT() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": u.Username,
			"exp":      time.Now().Add(time.Hour * 876000).Unix(), // 100 years
		})

	tokenString, err := token.SignedString(config.GetSharedConfig().JwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Validate checks the struct fields.
func (u *User) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

func GetUserByToken(tokenString string) (*User, error) {

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetSharedConfig().JwtSecret), nil
	})

	// Extract user username from claims
	username := (claims)["username"].(string)
	exp, _ := (claims)["exp"].(float64)
	if int64(exp) < time.Now().Unix() {
		return nil, fmt.Errorf("Token Expired")
	}

	usr, err := GetUserByUsername(username)

	return &usr, err
}

func GetUserByUsername(username string) (User, error) {

	usr := User{}
	err := database.Collection(config.GetConfigDB().UserColl).FindOne(context.Background(), bson.M{"username": username}).Decode(&usr)
	if err != nil {
		return usr, err
	}

	return usr, nil
}

func (u *User) AddUserToDB() error {

	collection := database.Collection(config.GetConfigDB().UserColl)
	// check if exists
	var usr User
	err := collection.FindOne(context.Background(), bson.M{"username": u.Username}).Decode(&usr)
	if err == nil {
		return nil
	}

	// Add user
	_, err = collection.InsertOne(context.Background(), u)
	if err != nil {
		return err
	}

	return nil
}
