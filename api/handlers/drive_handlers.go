package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/poriamsz55/distork/api/models/file"
	"github.com/poriamsz55/distork/api/models/user"
	config "github.com/poriamsz55/distork/configs"
	"github.com/poriamsz55/distork/database"
	"github.com/poriamsz55/distork/utils"
	"go.mongodb.org/mongo-driver/bson"
)

// Handler to upload files using streaming
func UploadFile(c echo.Context) error {

	usr := c.Get("user").(*user.User)

	currentPath := c.QueryParam("path")
	if currentPath == "" {
		currentPath = "." // Default to root if no path is provided
	}

	// Get form file
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	// Check the file size
	fileSize := file.Size
	newDriveUsed := usr.DriveUsed + fileSize

	// Check if new usage exceeds allowed drive size
	if newDriveUsed > usr.DriveSize {
		return c.String(http.StatusForbidden, "Insufficient drive space.")
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Create user-specific directory if not exists
	userDir := filepath.Join(config.GetConfigDrive().UploadDir, usr.Username, currentPath)
	if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
		return err
	}

	// Create destination file
	var dstPath string
	var copyCount int = 0
	for {
		dstPath = filepath.Join(userDir, utils.SanitizeFileName(file.Filename, copyCount))
		// Rename the destination file if it exists
		if _, err := os.Stat(dstPath); err == nil {
			copyCount++
		} else {
			break
		}
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Stream the uploaded file to the destination
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	collection := database.Collection(config.GetConfigDB().UserColl)

	// Update DriveUsed for the user
	usr.DriveUsed = newDriveUsed
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"username": usr.Username},
		bson.M{"$set": bson.M{"drive_used": usr.DriveUsed}},
	)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to update drive usage: %s", err))
	}

	return c.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully.", file.Filename))
}

func UploadFileChunk(c echo.Context) error {
	usr := c.Get("user").(*user.User)

	// Get chunk details
	chunkNumber := c.FormValue("chunkNumber") // The index of the current chunk
	totalChunks := c.FormValue("totalChunks") // Total number of chunks
	fileName := c.FormValue("fileName")       // The original file name
	timestamp := c.FormValue("timestamp")     // Timestamp from the client
	if timestamp == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing timestamp")
	}

	chunkFile, err := c.FormFile("file")
	if err != nil {
		return err
	}

	// Check the size of the chunk
	chunkSize := chunkFile.Size
	newDriveUsed := usr.DriveUsed + chunkSize

	// Check if new usage exceeds allowed drive size
	if newDriveUsed > usr.DriveSize {
		return c.String(http.StatusForbidden, "Insufficient drive space.")
	}

	currentPath := c.QueryParam("path")
	if currentPath == "" {
		currentPath = "." // Default to root if no path is provided
	}

	// Create user-specific directory if not exists
	userDir := filepath.Join(config.GetConfigDrive().UploadDir, usr.Username, currentPath)
	if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
		return err
	}

	// Open the chunk file
	src, err := chunkFile.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Save the chunk with a name including the timestamp
	chunkFileName := fmt.Sprintf("%s.part-%s-%s", utils.SanitizeFileName(fileName, 0), chunkNumber, timestamp)
	chunkPath := filepath.Join(userDir, chunkFileName)
	dst, err := os.Create(chunkPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the chunk to the destination
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	// Check if all chunks are uploaded
	currentChunk, _ := strconv.Atoi(chunkNumber)
	total, _ := strconv.Atoi(totalChunks)

	if currentChunk == total-1 {
		// Once all chunks are uploaded, combine them into the final file

		// Determine the final file name, handling duplicates
		var finalFileName string
		var copyCount int = 0
		for {
			finalFileName = utils.SanitizeFileName(fileName, copyCount)
			finalDstPath := filepath.Join(userDir, finalFileName)
			if _, err := os.Stat(finalDstPath); err == nil {
				// File exists, increment copy count
				copyCount++
			} else {
				// File does not exist, break the loop
				break
			}
		}

		finalDstPath := filepath.Join(userDir, finalFileName)
		dst, err := os.Create(finalDstPath)
		if err != nil {
			return err
		}
		defer dst.Close()

		// Combine all the chunks
		for i := 0; i < total; i++ {
			chunkFileName := fmt.Sprintf("%s.part-%d-%s", utils.SanitizeFileName(fileName, 0), i, timestamp)
			chunkPath := filepath.Join(userDir, chunkFileName)
			part, err := os.Open(chunkPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(dst, part); err != nil {
				part.Close()
				return err
			}
			part.Close()
			os.Remove(chunkPath) // Remove the chunk files after combining
		}

		collection := database.Collection(config.GetConfigDB().UserColl)

		// Update DriveUsed for the user after final file creation
		usr.DriveUsed = newDriveUsed
		_, err = collection.UpdateOne(
			context.Background(),
			bson.M{"username": usr.Username},
			bson.M{"$set": bson.M{"drive_used": usr.DriveUsed}},
		)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to update drive usage: %s", err))
		}

		// Optionally, you can inform the user about the final file name
		return c.JSON(http.StatusOK, map[string]string{
			"message":  fmt.Sprintf("File uploaded successfully as %s.", finalFileName),
			"fileName": finalFileName,
		})
	}

	return c.String(http.StatusOK, fmt.Sprintf("Chunk %d uploaded successfully.", currentChunk))
}

// Handler to list files and folders for a specific user
func ListFilesAndFolders(c echo.Context) error {
	usr := c.Get("user").(*user.User) // Get the authenticated user

	// Get the current path from the query parameter
	currentPath := c.QueryParam("path")
	if currentPath == "" {
		currentPath = "." // Default to root if no path is provided
	}

	uploadPath := ""
	uploadBase := ""
	if usr.Role == config.RoleAdmin {
		uploadPath = filepath.Join(config.GetConfigDrive().UploadDir, currentPath)
		uploadBase = config.GetConfigDrive().UploadDir
	} else {
		uploadPath = filepath.Join(config.GetConfigDrive().UploadDir, usr.Username, currentPath)
		uploadBase = filepath.Join(config.GetConfigDrive().UploadDir, usr.Username)
	}

	// Ensure the directory exists
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Directory does not exist",
		})
	}

	// Read the directory and return the list of files and folders
	fileList := []file.File{}
	entries, err := os.ReadDir(uploadPath) // Read the directory contents
	if err != nil {
		return err
	}

	for _, entry := range entries {
		info, err := entry.Info() // Get file information
		if err != nil {
			return err
		}

		// TODO: use info.Name() for Filename
		fileList = append(fileList, file.File{
			Filename:  info.Name(),
			UUsername: usr.Username,
			IsDir:     info.IsDir(),
			Path:      strings.TrimPrefix(filepath.Join(uploadBase, info.Name()), uploadBase), // Trim the base path for relative paths
			Size:      info.Size(),
			ModTime:   info.ModTime(),
		})
	}

	// Sort files by modification time (most recent first)
	sort.Slice(fileList, func(i, j int) bool {
		return fileList[i].ModTime.After(fileList[j].ModTime)
	})

	if len(fileList) == 0 {
		return c.JSON(http.StatusOK, nil)
	}

	return c.JSON(http.StatusOK, fileList)
}

// Handler to download a file using streaming for a specific user
func DownloadFile(c echo.Context) error {
	// Get the authenticated user
	usr := c.Get("user").(*user.User)

	// Get the requested file path relative to the user's root directory
	// This can be a relative path (e.g., subfolder/file.txt)
	requestedFile := c.QueryParam("path")
	if requestedFile == "" {
		requestedFile = "." // Default to root if no path is provided
	}

	// Sanitize the filename
	safeFilename, err := url.QueryUnescape(requestedFile)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid filename: %s", err))
	}

	// Construct the full path to the file within the user's directory
	userDir := filepath.Join(config.GetConfigDrive().UploadDir, usr.Username)
	safePath := filepath.Join(userDir, safeFilename)

	// Ensure that the requested file path is inside the user's directory
	if !strings.HasPrefix(safePath, userDir) {
		return c.String(http.StatusForbidden, "Invalid file path")
	}

	// Open the file for downloading
	file, err := os.Open(safePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.String(http.StatusNotFound, "File not found")
		}
		return err
	}
	defer file.Close()

	// Set headers for downloading the file
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(safeFilename)))
	c.Response().Header().Set(echo.HeaderContentType, "application/octet-stream")
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")

	// Stream the file to the response
	if _, err = io.Copy(c.Response().Writer, file); err != nil {
		return err
	}

	return nil
}

// Handler to delete a file for a specific user
func DeleteFile(c echo.Context) error {
	// Get the authenticated user
	usr := c.Get("user").(*user.User)

	// Retrieve the URL-encoded filename from the URL parameter
	encodedFilename := c.QueryParam("path")
	if encodedFilename == "" {
		encodedFilename = "." // Default to root if no path is provided
	}

	// Decode the filename to handle spaces encoded as %20
	filename, err := url.QueryUnescape(encodedFilename)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid filename: %s", err))
	}

	// Construct the full path to the file within the user's directory
	userDir := filepath.Join(config.GetConfigDrive().UploadDir, usr.Username)
	safePath := filepath.Join(userDir, filename)

	// Ensure that the requested file path is inside the user's directory
	if !strings.HasPrefix(safePath, userDir) {
		return c.String(http.StatusForbidden, "Invalid file path")
	}

	// Get the file info to determine its size
	fileInfo, err := os.Stat(safePath)
	if os.IsNotExist(err) {
		return c.String(http.StatusNotFound, fmt.Sprintf("File %s not found.", filename))
	} else if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error retrieving file info: %s", err))
	}

	// Update DriveUsed for the user
	fileSize := fileInfo.Size() // Size of the deleted file
	usr.DriveUsed -= fileSize   // Decrease the used space

	collection := database.Collection(config.GetConfigDB().UserColl)

	// Update user's DriveUsed in the database
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"username": usr.Username},
		bson.M{"$set": bson.M{"drive_used": usr.DriveUsed}},
	)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to update drive usage: %s", err))
	}

	// Remove the file
	if err := os.Remove(safePath); err != nil {
		if os.IsNotExist(err) {
			return c.String(http.StatusNotFound, fmt.Sprintf("File %s not found.", filename))
		}
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to delete file: %s", err))
	}

	return c.String(http.StatusOK, fmt.Sprintf("File %s deleted successfully!", filename))
}
