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
	config Config
}

// New creates a new WebDAV server
func New(config Config) *WebDAVServer {
	// Initialize WebDAV root directory
	if err := os.MkdirAll(config.RootDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create WebDAV root directory: %v", err))
	}

	r := gin.Default()

	server := &WebDAVServer{
		Engine: r,
		config: config,
	}

	// Set up routes
	server.SetupRoutes()

	return server
}

// SetupRoutes configures the WebDAV routes
func (s *WebDAVServer) SetupRoutes() {
	// Add middleware to handle WebDAV preflight requests
	s.Use(func(c *gin.Context) {
		if c.Request.Method == "PROPFIND" {
			c.Header("DAV", "1")
			c.Header("Content-Type", "application/xml; charset=utf-8")
		}
		c.Next()
	})
	
	// Explicitly register PROPFIND method with proper handling
	s.Handle("PROPFIND", "/*path", func(c *gin.Context) {
		path := c.Param("path")
		fullPath := filepath.Join(s.config.RootDir, path)
		s.handlePropfind(c, fullPath)
	})
	
	// Handle other WebDAV methods
	s.Any("/*path", func(c *gin.Context) {
		path := c.Param("path")
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
	log.Printf("WebDAV server starting on port %s", s.config.Port)
	log.Printf("WebDAV root directory: %s", s.config.RootDir)
	if err := s.Run(":" + s.config.Port); err != nil {
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
	XMLName  xml.Name `xml:"multistatus"`
	Response []Response `xml:"response"`
}

type Response struct {
	Href  string `xml:"href"`
	Prop  Prop   `xml:"propstat>prop"`
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
	ms.Response = make([]Response, 1)
	ms.Response[0].Href = c.Request.URL.Path
	ms.Response[0].Status = "HTTP/1.1 200 OK"

	// Set properties
	ms.Response[0].Prop.Name = filepath.Base(fullPath)
	ms.Response[0].Prop.DisplayName = filepath.Base(fullPath)
	ms.Response[0].Prop.IsCollection = "0"
	ms.Response[0].Prop.ContentType = "application/octet-stream"
	ms.Response[0].Prop.Length = "0"
	ms.Response[0].Prop.LastModified = ""

	if fileInfo != nil {
		if fileInfo.IsDir() {
			ms.Response[0].Prop.IsCollection = "1"
			ms.Response[0].Prop.ContentType = "httpd/unix-directory"
		}
		ms.Response[0].Prop.Length = fmt.Sprintf("%d", fileInfo.Size())
		ms.Response[0].Prop.LastModified = fileInfo.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	}

	// Log response
	log.Printf("Sending PROPFIND response for %s", fullPath)

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
