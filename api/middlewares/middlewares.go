package middlewares

import (
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func MiddleWares(e *echo.Group) {

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// e.Use(middleware.BodyLimit("4G")) // Increase the request body size limit for large files

	// Sample Go code using the Echo framework to handle CORS preflight requests
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*", "https://192.168.1.14:3000"}, // Adjust this to match your client's origin
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST,
			echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType,
			echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Setup Logger
	logFile, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatal("Failed to log to file, using default stderr")
	}
	logrus.SetOutput(logFile)

	// Middleware to log HTTP requests
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: logFile,
	}))

	logrus.SetOutput(&lumberjack.Logger{
		Filename:   "logs.txt",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	})
}

func UploadMiddleWares(e *echo.Group) {

	// Check numer of requests for each IP
	//
	InitIPRateLimiter()

	//
	limiterConfig := RequestLimiterConfig{
		RequestLimit: 200,            // Max 200 requests
		Interval:     time.Minute,    // per minute
		BlockTime:    24 * time.Hour, // Block for 24 hours
	}

	e.Use(RateLimitMiddleware(limiterConfig))

}
