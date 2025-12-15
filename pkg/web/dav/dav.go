package dav

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/cgang/file-hub/pkg/stor"
	"github.com/cgang/file-hub/pkg/web/auth"
	"github.com/gin-gonic/gin"
)

const (
	davNamespace = "DAV:"
)

// PropfindRequest represents the XML request body for PROPFIND
type PropfindRequest struct {
	XMLName  xml.Name      `xml:"DAV: propfind"`
	AllProp  *struct{}     `xml:"DAV: allprop,omitempty"`
	PropName *struct{}     `xml:"DAV: propname,omitempty"`
	Prop     *PropfindProp `xml:"DAV: prop,omitempty"`
}

// PropfindProp represents the prop element in PROPFIND request
type PropfindProp struct {
	DisplayName   *struct{} `xml:"DAV: displayname,omitempty"`
	ResourceType  *struct{} `xml:"DAV: resourcetype,omitempty"`
	ContentType   *struct{} `xml:"DAV: getcontenttype,omitempty"`
	ContentLength *struct{} `xml:"DAV: getcontentlength,omitempty"`
	LastModified  *struct{} `xml:"DAV: getlastmodified,omitempty"`
	CreationDate  *struct{} `xml:"DAV: creationdate,omitempty"`
	ETag          *struct{} `xml:"DAV: getetag,omitempty"`
	// Add more properties as needed
}

func setDavHeaders(c *gin.Context) {
	c.Header("DAV", "1")
	c.Header("MS-Author-Via", "DAV")
	//c.Header("Content-Type", "text/xml; charset=utf-8")
}

// getAuthenticatedUser retrieves the authenticated user from the context
func getAuthenticatedUser(c *gin.Context) (*model.User, error) {
	userValue, exists := c.Get("user")
	if !exists {
		return nil, fmt.Errorf("no authenticated user found")
	}

	user, ok := userValue.(*model.User)
	if !ok {
		return nil, fmt.Errorf("invalid user type in context")
	}

	return user, nil
}

// Register configures the WebDAV routes
func Register(v1 *gin.RouterGroup) {
	// Note: Authentication should be handled by a middleware in the calling code
	v1.Use(setDavHeaders)

	v1.OPTIONS("/:repo/*path", handleOptions)

	v1.Use(auth.Authenticate)

	v1.PUT("/:repo/*path", handlePut)
	v1.DELETE("/:repo/*path", handleDelete)
	v1.GET("/:repo/*path", handleGet)

	v1.Handle("PROPFIND", "/:repo/*path", handlePropfind)
	v1.Handle("MKCOL", "/:repo/*path", handleMkcol)
	v1.Handle("COPY", "/:repo/*path", handleCopyMove)
	v1.Handle("MOVE", "/:repo/*path", handleCopyMove)
}

type Prop struct {
	DisplayName  string        `xml:"D:displayname"`
	ResourceType *ResourceType `xml:"D:resourcetype,omitempty"`
	ContentType  string        `xml:"D:getcontenttype"`
	Length       string        `xml:"D:getcontentlength,omitempty"`
	LastModified string        `xml:"D:getlastmodified"`
	CreationDate string        `xml:"D:creationdate,omitempty"`
	ETag         string        `xml:"D:getetag,omitempty"`
}

type ResourceType struct {
	XmlData string `xml:",innerxml"`
}

func CollectionType() *ResourceType {
	return &ResourceType{XmlData: "<D:collection/>"}
}

type Multistatus struct {
	XMLName  xml.Name   `xml:"D:multistatus"`
	DavNS    string     `xml:"xmlns:D,attr"`
	Response []Response `xml:"D:response"`
}

type Response struct {
	Href     string   `xml:"D:href"`
	Propstat Propstat `xml:"D:propstat"`
}

type Propstat struct {
	Prop   Prop   `xml:"D:prop"`
	Status string `xml:"D:status"`
}

// ErrorBody is used for WebDAV error responses
type ErrorBody struct {
	XMLName xml.Name `xml:"D:error"`
	DavNS   string   `xml:"xmlns:dav,attr"`
	Message string   `xml:",innerxml"`
}

// sendError sends a standardized WebDAV error response
func sendError(c *gin.Context, status int, format string, a ...any) {
	c.XML(status, &ErrorBody{
		XMLName: xml.Name{Space: "DAV", Local: "error"},
		DavNS:   davNamespace,
		Message: fmt.Sprintf(format, a...),
	})
}

func getResource(c *gin.Context) (*model.Resource, error) {
	name := c.Param("repo")
	repo, err := stor.GetRepository(c, name)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Repository not found")
		return nil, fmt.Errorf("get repository %s failed: %w", name, err)
	}

	return &model.Resource{
		Repo: repo,
		Path: strings.TrimSuffix(c.Param("path"), "/"),
	}, nil
}

// getResourceByUrl parses a URL and returns the corresponding Resource
func getResourceByUrl(ctx context.Context, urlStr string) (*model.Resource, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Extract the path and remove the leading slash
	path := strings.TrimPrefix(u.Path, "/")

	// Split the path to get repo and file path
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid path: %s", path)
	}

	base := parts[0]
	name := parts[1]

	r, err := stor.GetRepository(ctx, base)
	if err != nil {
		return nil, err
	}

	return &model.Resource{Repo: r, Path: name}, nil
}

// handlePropfind handles PROPFIND requests
func handlePropfind(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	// Parse the PROPFIND request using c.Bind
	propfindReq := &PropfindRequest{}
	if c.Request.ContentLength > 0 {
		if err := c.Bind(propfindReq); err != nil {
			sendError(c, http.StatusBadRequest, "Failed to parse XML: %v", err)
			return
		}
	} else {
		// If no body, default to allprop as per RFC
		propfindReq.AllProp = &struct{}{}
	}

	// Handle Depth header
	depth := c.GetHeader("Depth")

	// We only support Depth: 0 and Depth: 1
	switch depth {
	case "0", "1":
		// Valid depth values, continue processing
	default:
		sendError(c, http.StatusForbidden, "Depth %s is not supported", depth)
		return
	}

	// Log request
	log.Printf("Handling PROPFIND request for %s with depth %s", resource, depth)

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionRead); err != nil {
		log.Printf("Permission denied for %s: %v", resource, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	// Build response
	ms := &Multistatus{DavNS: davNamespace}

	// Get file info using storage abstraction
	file, err := stor.GetFileInfo(c, resource)
	if err != nil {
		if stor.IsNotFound(err) {
			sendError(c, http.StatusNotFound, "File not found")
			return
		}

		log.Printf("Error accessing file %s: %v", resource, err)
		sendError(c, http.StatusInternalServerError, "Error accessing file: %v", err)
		return
	}

	// Add the file/directory itself
	ms.Response = append(ms.Response, CreateResponse(c.Request.URL.Path, file, propfindReq))

	// If depth is 1 and it's a directory, list its contents
	if depth == "1" && file.IsDir {
		files, err := stor.ListDir(c, resource.Repo, file)
		if err != nil {
			log.Printf("Error reading directory %s: %v", resource, err)
			sendError(c, http.StatusInternalServerError, "Failed to read directory: %v", err)
			return
		}

		for _, entry := range files {
			// Construct proper href for each entry
			entryHref := path.Join(c.Request.URL.Path, path.Base(entry.Path))
			if entry.IsDir && !strings.HasSuffix(entryHref, "/") {
				entryHref += "/"
			}
			ms.Response = append(ms.Response, CreateResponse(entryHref, entry, propfindReq))
		}
	}

	XML(c, http.StatusMultiStatus, ms)
}

func XML(c *gin.Context, code int, body any) {
	var buf bytes.Buffer
	fmt.Fprint(&buf, xml.Header)
	if err := xml.NewEncoder(&buf).Encode(body); err != nil {
		log.Printf("Failed to encode XML: %s", err)
	}

	c.Data(code, "text/xml; charset=utf-8", buf.Bytes())
}

// CreateResponse creates a WebDAV response with proper properties (exported for testing)
func CreateResponse(href string, file *model.FileObject, req *PropfindRequest) Response {
	name := path.Base(file.Path)
	if name == "." || name == "/" {
		name = ""
	}

	prop := Prop{}

	// Determine which properties to include based on the request
	if req.PropName != nil {
		// For propname requests, we only return the property names, not values
		// We'll handle this by sending an empty multistatus response with status 200
		// and another propstat with status 404 for properties not found
		// For simplicity, we'll just return an empty prop
	} else if req.AllProp != nil || req.Prop == nil {
		// allprop request or no specific properties requested
		prop.DisplayName = name
		prop.LastModified = file.ModTime.UTC().Format(time.RFC1123)
		prop.CreationDate = file.ModTime.Format(time.RFC3339)

		if file.IsDir {
			// For directories, specify the resourcetype as collection
			prop.ResourceType = CollectionType()
			prop.ContentType = "httpd/unix-directory"
			if !strings.HasSuffix(href, "/") {
				href = href + "/"
			}
		} else {
			// For files, leave resourcetype empty and specify content type and length
			prop.ResourceType = nil
			prop.ContentType = file.ContentType()
			prop.Length = fmt.Sprintf("%d", file.Size)
			// Generate a simple etag based on modtime and size
			prop.ETag = fmt.Sprintf("%x-%x", file.ModTime.Unix(), file.Size)
		}
	} else {
		// Specific properties requested
		if req.Prop.DisplayName != nil {
			prop.DisplayName = name
		}
		if req.Prop.LastModified != nil {
			prop.LastModified = file.ModTime.UTC().Format(time.RFC1123)
		}
		if req.Prop.CreationDate != nil {
			prop.CreationDate = file.ModTime.Format(time.RFC3339)
		}
		if req.Prop.ResourceType != nil {
			if file.IsDir {
				prop.ResourceType = CollectionType()
			} else {
				prop.ResourceType = nil
			}
		}
		if req.Prop.ContentType != nil {
			if file.IsDir {
				prop.ContentType = "httpd/unix-directory"
			} else {
				prop.ContentType = file.ContentType()
			}
		}
		if req.Prop.ContentLength != nil && !file.IsDir {
			prop.Length = fmt.Sprintf("%d", file.Size)
		}
		if req.Prop.ETag != nil && !file.IsDir {
			prop.ETag = fmt.Sprintf("%x-%x", file.ModTime.Unix(), file.Size)
		}
	}

	u := &url.URL{Path: href}

	return Response{
		Href: u.EscapedPath(),
		Propstat: Propstat{
			Prop:   prop,
			Status: "HTTP/1.1 200 OK",
		},
	}
}

// handlePut handles PUT requests
func handlePut(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionWrite); err != nil {
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	// Write file using storage abstraction
	if err := stor.PutFile(c, resource, c.Request.Body); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to write file: %v", err)
		return
	}

	c.Status(http.StatusCreated)
}

// handleDelete handles DELETE requests
func handleDelete(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionDelete); err != nil {
		log.Printf("Permission denied for %s: %v", resource, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	if err := stor.DeleteFile(c, resource); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to delete file: %v", err)
		return
	}
	c.Status(http.StatusNoContent)
}

// handleMkcol handles MKCOL requests
func handleMkcol(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionWrite); err != nil {
		log.Printf("Permission denied for %s: %v", resource, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	if err := stor.CreateDir(c, resource); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to create directory: %v", err)
		return
	}
	c.Status(http.StatusCreated)
}

// handleCopyMove handles COPY and MOVE requests
func handleCopyMove(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	// Parse destination path
	destination := c.Request.Header.Get("Destination")
	destRes, err := getResourceByUrl(c.Request.Context(), destination)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid destination: %s", err)
		return
	}

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionRead); err != nil {
		sendError(c, http.StatusForbidden, "Permission denied for source")
		return
	}

	if err := stor.CheckPermission(c, user.ID, destRes, stor.PermissionWrite); err != nil {
		sendError(c, http.StatusForbidden, "Permission denied for destination")
		return
	}

	// Handle COPY or MOVE
	if c.Request.Method == "COPY" {
		// Copy file/directory using storage
		if err := stor.CopyFile(c, resource, destRes); err != nil {
			sendError(c, http.StatusInternalServerError, "Failed to copy file: %v", err)
			return
		}
	} else {
		// Move file/directory using storage
		if err := stor.MoveFile(c, resource, destRes); err != nil {
			sendError(c, http.StatusInternalServerError, "Failed to move file: %v", err)
			return
		}
	}

	c.Status(http.StatusCreated)
}

// handleGet handles GET requests
func handleGet(c *gin.Context) {
	// Get authenticated user
	user, err := getAuthenticatedUser(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	resource, err := getResource(c)
	if err != nil {
		return
	}

	if err := stor.CheckPermission(c, user.ID, resource, stor.PermissionRead); err != nil {
		log.Printf("Permission denied for %s: %v", resource, err)
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	info, err := stor.GetFileInfo(c, resource)
	if err != nil {
		if os.IsNotExist(err) {
			sendError(c, http.StatusNotFound, "File not found")
			return
		}
		sendError(c, http.StatusInternalServerError, "Error accessing file: %v", err)
		return
	}

	if info.IsDir {
		sendError(c, http.StatusBadRequest, "Cannot GET a directory")
		return
	}

	c.Header("Content-Type", info.ContentType())
	c.Header("Content-Length", fmt.Sprintf("%d", info.Size))

	file, err := stor.OpenFile(c, resource)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Error opening file: %v", err)
		return
	}
	defer file.Close()

	if _, err := io.Copy(c.Writer, file); err != nil {
		log.Printf("Failed to copy file content: %s", err)
	}
}

func handleOptions(c *gin.Context) {
	c.Header("Allow", "OPTIONS,GET,POST,PUT,DELETE,COPY,MOVE,PROPFIND,MKCOL,LOCK,UNLOCK")
	c.Status(http.StatusOK)
}
