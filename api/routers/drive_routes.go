package router

import (
	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/handlers"
)

func UploadRoutes(e *echo.Group) {
	e.POST("", handlers.UploadFileChunk)
	e.POST("/upload", handlers.UploadFile)
}

func DriveRoutes(e *echo.Group) {
	e.GET("/files", handlers.ListFilesAndFolders)
	e.GET("/download", handlers.DownloadFile)
	e.GET("/delete", handlers.DeleteFile)
}
