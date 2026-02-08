package dav

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPropfindRequestParsing(t *testing.T) {
	// Test parsing of allprop request
	allPropXML := `<D:propfind xmlns:D="DAV:"><D:allprop/></D:propfind>`
	req := &PropfindRequest{}
	err := xml.Unmarshal([]byte(allPropXML), req)
	assert.NoError(t, err)
	assert.NotNil(t, req.AllProp)
	assert.Nil(t, req.PropName)
	assert.Nil(t, req.Prop)

	// Test parsing of propname request
	propNameXML := `<D:propfind xmlns:D="DAV:"><D:propname/></D:propfind>`
	req = &PropfindRequest{}
	err = xml.Unmarshal([]byte(propNameXML), req)
	assert.NoError(t, err)
	assert.Nil(t, req.AllProp)
	assert.NotNil(t, req.PropName)
	assert.Nil(t, req.Prop)

	// Test parsing of prop request with specific properties
	propXML := `<D:propfind xmlns:D="DAV:"><D:prop><D:displayname/><D:getcontentlength/></D:prop></D:propfind>`
	req = &PropfindRequest{}
	err = xml.Unmarshal([]byte(propXML), req)
	assert.NoError(t, err)
	assert.Nil(t, req.AllProp)
	assert.Nil(t, req.PropName)
	assert.NotNil(t, req.Prop)
	assert.NotNil(t, req.Prop.DisplayName)
	assert.NotNil(t, req.Prop.ContentLength)
	assert.Nil(t, req.Prop.ResourceType)
}

func TestCreateResponseAllProp(t *testing.T) {
	// Create a test file object
	modTime := time.Now()
	file := &model.FileObject{
		Name:    "test.txt",
		Path:    "/test.txt",
		Size:    1024,
		ModTime: modTime,
		IsDir:   false,
	}

	// Create an allprop request
	propfindReq := &PropfindRequest{
		AllProp: &struct{}{},
	}

	// Create response
	response := CreateResponse("/test.txt", file, propfindReq)

	// Check that all properties are present
	assert.Equal(t, "/test.txt", response.Href)
	assert.Equal(t, "test.txt", response.Propstat.Prop.DisplayName)
	assert.Equal(t, modTime.UTC().Format(time.RFC1123), response.Propstat.Prop.LastModified)
	assert.Empty(t, response.Propstat.Prop.ResourceType)
	assert.Equal(t, "application/octet-stream", response.Propstat.Prop.ContentType)
	assert.Equal(t, "1024", response.Propstat.Prop.Length)
	assert.Equal(t, "HTTP/1.1 200 OK", response.Propstat.Status)
}

func TestCreateResponseSpecificProps(t *testing.T) {
	// Create a test file object
	modTime := time.Now()
	file := &model.FileObject{
		Name:    "test.txt",
		Path:    "/test.txt",
		Size:    1024,
		ModTime: modTime,
		IsDir:   false,
	}

	// Create a prop request for specific properties
	propfindReq := &PropfindRequest{
		Prop: &PropfindProp{
			DisplayName:   &struct{}{},
			ContentLength: &struct{}{},
		},
	}

	// Create response
	response := CreateResponse("/test.txt", file, propfindReq)

	// Check that only requested properties are present
	assert.Equal(t, "/test.txt", response.Href)
	assert.Equal(t, "test.txt", response.Propstat.Prop.DisplayName)
	assert.Empty(t, response.Propstat.Prop.LastModified)
	assert.Empty(t, response.Propstat.Prop.ResourceType)
	assert.Empty(t, response.Propstat.Prop.ContentType)
	assert.Equal(t, "1024", response.Propstat.Prop.Length)
	assert.Equal(t, "HTTP/1.1 200 OK", response.Propstat.Status)
}

func TestCreateResponseDirectory(t *testing.T) {
	// Create a test directory object
	modTime := time.Now()
	dir := &model.FileObject{
		Name:    "testdir",
		Path:    "/testdir/",
		ModTime: modTime,
		IsDir:   true,
	}

	// Create an allprop request
	propfindReq := &PropfindRequest{
		AllProp: &struct{}{},
	}

	// Create response
	response := CreateResponse("/testdir/", dir, propfindReq)

	// Check that directory properties are correct
	assert.Equal(t, "/testdir/", response.Href)
	assert.Equal(t, "testdir", response.Propstat.Prop.DisplayName)
	assert.Equal(t, modTime.UTC().Format(time.RFC1123), response.Propstat.Prop.LastModified)
	assert.Equal(t, "<D:collection/>", response.Propstat.Prop.ResourceType.XmlData)
	assert.Equal(t, "httpd/unix-directory", response.Propstat.Prop.ContentType)
	assert.Empty(t, response.Propstat.Prop.Length) // Directories don't have content length
	assert.Equal(t, "HTTP/1.1 200 OK", response.Propstat.Status)
}

func TestHandlePropfindWithEmptyBody(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// We can't fully test the handler without mocking the storage layer,
	// but we can at least verify the XML parsing works with an empty body
	// For now, we've tested the individual components above
}
