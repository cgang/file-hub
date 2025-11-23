package webdav

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cgang/file-hub/internal/config"
	"github.com/cgang/file-hub/internal/stor"
	"github.com/gin-gonic/gin"
)

// Config holds WebDAV server configuration
type Config struct {
	RootDir string
	Port    string
}

// WebDAVServer represents the WebDAV server
type WebDAVServer struct {
	*gin.Engine
	config  Config
	storage stor.Storage
}

// New creates a new WebDAV server
func New(config Config, storage stor.Storage) *WebDAVServer {
	// Initialize WebDAV root directory
	if err := storage.CreateDir(config.RootDir); err != nil {
		panic(fmt.Sprintf("Failed to create WebDAV root directory: %v", err))
	}

	r := gin.Default()

	server := &WebDAVServer{
		Engine:  r,
		config:  config,
		storage: storage,
	}

	// Set up routes
	server.SetupRoutes()

	return server
}

// NewFromConfig creates a new WebDAV server from the new configuration structure
func NewFromConfig(webdavConfig config.WebDAVConfig, storage stor.Storage, storageRootDir string) *WebDAVServer {
	config := Config{
		RootDir: storageRootDir,
		Port:    webdavConfig.Port,
	}
	return New(config, storage)
}

// SetupRoutes configures the WebDAV routes
func (s *WebDAVServer) SetupRoutes() {
	// Create a group for WebDAV routes to avoid conflicts with frontend
	v1 := s.Group("/webdav")
	v1.Use(func(c *gin.Context) {
		// Add WebDAV headers
		if c.Request.Method == "PROPFIND" {
			c.Header("DAV", "1")
			c.Header("Content-Type", "application/xml; charset=utf-8")
		}

		c.Next()
	})

	// Explicitly register PROPFIND method with proper handling
	v1.Handle("PROPFIND", "/*path", func(c *gin.Context) {
		path := strings.TrimPrefix(c.Param("path"), "/webdav")
		fullPath := filepath.Join(s.config.RootDir, path)
		s.handlePropfind(c, fullPath)
	})

	// Handle other WebDAV methods
	v1.Any("/*path", func(c *gin.Context) {
		path := strings.TrimPrefix(c.Param("path"), "/webdav")
		fullPath := filepath.Join(s.config.RootDir, path)

		// Handle WebDAV methods
		switch c.Request.Method {
		case "PROPFIND":
			s.handlePropfind(c, fullPath)
		case "PUT":
			s.handlePut(c, fullPath)
		case "DELETE":
			s.handleDelete(c, fullPath)
		case "MKCOL":
			s.handleMkcol(c, fullPath)
		case "COPY", "MOVE":
			s.handleCopyMove(c, fullPath)
		case "GET":
			s.handleGet(c, fullPath)
		default:
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{
				"error": fmt.Sprintf("Method %s not supported", c.Request.Method),
			})
		}
	})
}

// Start starts the WebDAV server
func (s *WebDAVServer) Start() {
	addr := ":" + s.config.Port
	log.Printf("WebDAV server starting on %s", addr)
	log.Printf("WebDAV root directory: %s", s.config.RootDir)
	if err := s.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// XML structures for WebDAV
type Propfind struct {
	XMLName xml.Name `xml:"DAV: propfind"`
	Prop    Prop     `xml:"prop"`
}

type Prop struct {
	Name         string `xml:"name"`
	DisplayName  string `xml:"displayname"`
	IsCollection string `xml:"iscollection"`
	ContentType  string `xml:"getcontenttype"`
	Length       string `xml:"getcontentlength"`
	LastModified string `xml:"getlastmodified"`
}

type Multistatus struct {
	XMLName  xml.Name   `xml:"multistatus"`
	Response []Response `xml:"response"`
}

type Response struct {
	Href   string `xml:"href"`
	Prop   Prop   `xml:"propstat>prop"`
	Status string `xml:"status"`
}

type ErrorBody struct {
	XMLName xml.Name `xml:"error"`
	Message string   `xml:",innerxml"`
}

// handlePropfind handles PROPFIND requests
func (s *WebDAVServer) handlePropfind(c *gin.Context, fullPath string) {
	// Log request
	log.Printf("Handling PROPFIND request for %s", fullPath)

	// Check if file exists
	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		log.Printf("File not found: %s", fullPath)
		// Return proper XML error response
		c.XML(http.StatusNotFound, &ErrorBody{
			XMLName: xml.Name{Space: "DAV", Local: "error"},
			Message: "File not found",
		})
		return
	} else if err != nil {
		log.Printf("Error accessing file %s: %v", fullPath, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Build response
	var ms Multistatus

	// If it's a directory, list its contents
	if fileInfo.IsDir() {
		// Add the directory itself as the first response
		dirResponse := Response{
			Href:   c.Request.URL.Path,
			Status: "HTTP/1.1 200 OK",
		}
		dirResponse.Prop.Name = filepath.Base(fullPath)
		dirResponse.Prop.DisplayName = filepath.Base(fullPath)
		dirResponse.Prop.IsCollection = "1"
		dirResponse.Prop.ContentType = "httpd/unix-directory"
		dirResponse.Prop.Length = fmt.Sprintf("%d", fileInfo.Size())
		dirResponse.Prop.LastModified = fileInfo.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
		ms.Response = append(ms.Response, dirResponse)

		// Read directory contents
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			log.Printf("Error reading directory %s: %v", fullPath, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Add each entry to the response
		for _, entry := range entries {
			entryPath := filepath.Join(fullPath, entry.Name())
			entryInfo, err := entry.Info()
			if err != nil {
				log.Printf("Error getting info for %s: %v", entryPath, err)
				continue // Skip this entry
			}

			entryUrlPath := strings.TrimSuffix(c.Request.URL.Path, "/") + "/" + entry.Name()
			entryResponse := Response{
				Href:   entryUrlPath,
				Status: "HTTP/1.1 200 OK",
			}
			entryResponse.Prop.Name = entry.Name()
			entryResponse.Prop.DisplayName = entry.Name()
			if entry.IsDir() {
				entryResponse.Prop.IsCollection = "1"
				entryResponse.Prop.ContentType = "httpd/unix-directory"
			} else {
				entryResponse.Prop.IsCollection = "0"
				entryResponse.Prop.ContentType = "application/octet-stream"
			}
			entryResponse.Prop.Length = fmt.Sprintf("%d", entryInfo.Size())
			entryResponse.Prop.LastModified = entryInfo.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
			ms.Response = append(ms.Response, entryResponse)
		}
	} else {
		// It's a file, respond with just this file's information
		fileResponse := Response{
			Href:   c.Request.URL.Path,
			Status: "HTTP/1.1 200 OK",
		}
		fileResponse.Prop.Name = filepath.Base(fullPath)
		fileResponse.Prop.DisplayName = filepath.Base(fullPath)
		fileResponse.Prop.IsCollection = "0"
		fileResponse.Prop.ContentType = "application/octet-stream"
		fileResponse.Prop.Length = fmt.Sprintf("%d", fileInfo.Size())
		fileResponse.Prop.LastModified = fileInfo.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
		ms.Response = append(ms.Response, fileResponse)
	}

	// Log response
	log.Printf("Sending PROPFIND response for %s with %d items", fullPath, len(ms.Response))

	// Set XML content type and send response
	c.XML(http.StatusOK, ms)
}

// handlePut handles PUT requests
func (s *WebDAVServer) handlePut(c *gin.Context, fullPath string) {
	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create directory",
		})
		return
	}

	// Open file for writing
	file, err := os.Create(fullPath)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create file",
		})
		return
	}
	defer file.Close()

	// Copy content from request
	if _, err := io.Copy(file, c.Request.Body); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to write file",
		})
		return
	}

	c.Status(http.StatusCreated)
}

// handleDelete handles DELETE requests
func (s *WebDAVServer) handleDelete(c *gin.Context, fullPath string) {
	if err := os.RemoveAll(fullPath); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file",
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// handleMkcol handles MKCOL requests
func (s *WebDAVServer) handleMkcol(c *gin.Context, fullPath string) {
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create directory",
		})
		return
	}
	c.Status(http.StatusCreated)
}

// handleCopyMove handles COPY and MOVE requests
func (s *WebDAVServer) handleCopyMove(c *gin.Context, fullPath string) {
	destination := c.Request.Header.Get("Destination")
	if destination == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Destination header required",
		})
		return
	}

	// Parse destination path
	destPath := strings.TrimPrefix(destination, "/")
	destFullPath := filepath.Join(s.config.RootDir, destPath)

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(destFullPath), 0755); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create destination directory",
		})
		return
	}

	// Handle COPY or MOVE
	if c.Request.Method == "COPY" {
		// Copy file/directory
		if err := copyPath(fullPath, destFullPath); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to copy file",
			})
			return
		}
	} else {
		// MOVE = copy + delete
		if err := copyPath(fullPath, destFullPath); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to move file",
			})
			return
		}
		if err := os.RemoveAll(fullPath); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete source file after move",
			})
			return
		}
	}

	c.Status(http.StatusCreated)
}

// handleGet handles GET requests
func (s *WebDAVServer) handleGet(c *gin.Context, fullPath string) {
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	http.ServeFile(c.Writer, c.Request, fullPath)
}

// copyPath copies a file or directory
func copyPath(src, dst string) error {
	fileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		// Handle directory copy
		return copyDir(src, dst)
	}
	// Handle file copy
	return copyFile(src, dst)
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read directory contents
	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each file/dir
	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		dstPath := filepath.Join(dst, file.Name())

		if file.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	return err
}
