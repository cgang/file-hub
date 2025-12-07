package stor

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/cgang/file-hub/pkg/config"
	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
)

type FileMeta struct {
	Name         string
	Path         string
	IsDir        bool
	Size         int64
	LastModified time.Time
}

func newFileMeta(fullname string, mt time.Time) *FileMeta {
	return &FileMeta{
		Name:         path.Base(fullname),
		Path:         fullname,
		LastModified: mt,
	}
}

func newDirMeta(fullname string, mt time.Time) *FileMeta {
	return &FileMeta{
		Name:         path.Base(fullname),
		Path:         fullname,
		IsDir:        true,
		LastModified: mt,
	}
}

func (m *FileMeta) toObject(repoID, ownerID, parentID int) *model.FileObject {
	return &model.FileObject{
		RepoID:    repoID,
		OwnerID:   ownerID,
		ParentID:  parentID,
		Name:      m.Name,
		Path:      m.Path,
		Size:      m.Size,
		IsDir:     m.IsDir,
		UpdatedAt: m.LastModified,
	}
}

var (
	rootDirs []string
)

func Init(ctx context.Context, cfg *config.Config) {
	if cfg.S3 != nil {
		s3Client = newS3Client(cfg.S3)
	}
	rootDirs = cfg.RootDir
}

// IsNotFound return true if err is something not found.
func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func isConfiguredRoot(root string) bool {
	for _, dir := range rootDirs {
		if t := strings.TrimPrefix(root, dir); t != root && strings.HasPrefix(t, "/") {
			return true
		}
	}
	return false
}

func ValidRoot(root string) bool {
	root = path.Clean(root)
	if !isConfiguredRoot(root) {
		return false
	}

	if s, err := os.Stat(root); err == nil {
		return s.IsDir()
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(root, 0755); err == nil {
			return true
		} else {
			log.Printf("Failed to create directory %s: %s", root, err)
		}
	} else {
		log.Printf("Failed to check directory %s: %s", root, err)
	}

	return false
}

// GetFileInfo retrieves file metadata from the database
func GetFileInfo(ctx context.Context, resource *model.Resource) (*model.FileObject, error) {
	return db.GetFile(ctx, resource.Repo.ID, resource.Path)
}

// ListDir lists the contents of a directory
func ListDir(ctx context.Context, repo *model.Repository, parent *model.FileObject) ([]*model.FileObject, error) {
	if !parent.IsDir {
		return nil, nil // return nil for non directory files
	}

	storage, err := getStorage(repo)
	if err != nil {
		return nil, err
	}

	objects, err := db.GetChildFiles(ctx, parent.ID)
	if err != nil {
		return nil, err
	}

	var changed []*model.FileObject
	for _, obj := range objects {
		if obj.MimeType != nil {
			continue
		}

		if ct, err := storage.GetContentType(ctx, repo.Name, obj.Path); err == nil {
			obj.MimeType = aws.String(ct)
		} else {
			obj.MimeType = aws.String(getContentType(path.Ext(obj.Name)))
		}
		changed = append(changed, obj)
	}

	if len(changed) > 0 {
		if err := db.UpdateContentType(ctx, changed); err != nil {
			log.Printf("Failed to update content type: %s", err)
		}
	}

	return objects, nil
}

// CreateDir creates a directory entry in the database
func CreateDir(ctx context.Context, resource *model.Resource) error {
	object := &model.FileObject{
		RepoID:  resource.Repo.ID,
		OwnerID: resource.Repo.OwnerID,
		Name:    path.Base(resource.Path),
		Path:    path.Clean(resource.Path),
		Size:    0,
		IsDir:   true,
	}

	return db.CreateFile(ctx, object)
}

// Storage defines the interface for file storage backends
type Storage interface {
	// PutFile uploads a file to the storage backend
	PutFile(ctx context.Context, repo, name string, data io.Reader) error
	// OpenFile opens a file for reading from the storage backend
	OpenFile(ctx context.Context, repo, name string) (io.ReadCloser, error)
	// DeleteFile deletes a file from the storage backend
	DeleteFile(ctx context.Context, repo, name string) error
	// CopyFile copies a file within the storage backend
	CopyFile(ctx context.Context, repo, srcName, destName string) error
	// Scan scanes existing objects of storage.
	Scan(ctx context.Context, repo string, visit func(*FileMeta) error) error
	// GetContentType returns content type of file
	GetContentType(ctx context.Context, repo, name string) (string, error)
}

// getStorage returns the appropriate Storage implementation based on the repository's Root URL
func getStorage(repo *model.Repository) (Storage, error) {
	u, err := url.Parse(repo.Root)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "s3":
		return &s3Storage{u.Host}, nil
	case "file", "":
		return &fsStorage{u.Path}, nil
	default:
		return nil, errors.New("unsupported storage scheme: " + u.Scheme)
	}
}

// PutFile uploads a file to the appropriate storage backend
func PutFile(ctx context.Context, res *model.Resource, dataReader io.Reader) error {
	storage, err := getStorage(res.Repo)
	if err != nil {
		return err
	}

	return storage.PutFile(ctx, res.Repo.Name, res.Path, dataReader)
}

// OpenFile opens a file for reading from the appropriate storage backend
func OpenFile(ctx context.Context, resource *model.Resource) (io.ReadCloser, error) {
	storage, err := getStorage(resource.Repo)
	if err != nil {
		return nil, err
	}

	return storage.OpenFile(ctx, resource.Repo.Name, resource.Path)
}

// DeleteFile deletes a file from the appropriate storage backend
func DeleteFile(ctx context.Context, resource *model.Resource) error {
	storage, err := getStorage(resource.Repo)
	if err != nil {
		return err
	}

	return storage.DeleteFile(ctx, resource.Repo.Name, resource.Path)
}

// CopyFile copies a file within the same repository in the appropriate storage backend
func CopyFile(ctx context.Context, srcResource *model.Resource, destResource *model.Resource) error {
	if srcResource.Repo.ID != destResource.Repo.ID {
		return errors.New("cross-repository copy not supported yet")
	}

	storage, err := getStorage(srcResource.Repo)
	if err != nil {
		return err
	}

	return storage.CopyFile(ctx, srcResource.Repo.Name, srcResource.Path, destResource.Path)
}

// MoveFile moves a file within the same repository in the appropriate storage backend
func MoveFile(ctx context.Context, srcResource *model.Resource, destResource *model.Resource) error {
	if srcResource.Repo.ID != destResource.Repo.ID {
		return errors.New("cross-repository move not supported yet")
	}

	storage, err := getStorage(srcResource.Repo)
	if err != nil {
		return err
	}

	err = storage.CopyFile(ctx, srcResource.Repo.Name, srcResource.Path, destResource.Path)
	if err != nil {
		return err
	}

	return storage.DeleteFile(ctx, srcResource.Repo.Name, srcResource.Path)
}

// ImportFiles imports existing files from storage location.
func ImportFiles(ctx context.Context, repo *model.Repository) error {
	storage, err := getStorage(repo)
	if err != nil {
		return err
	}

	return storage.Scan(ctx, repo.Name, func(fm *FileMeta) error {
		parent, err := db.GetFile(ctx, repo.ID, path.Dir(fm.Path))
		if err != nil {
			return err
		}

		object := fm.toObject(repo.ID, repo.OwnerID, parent.ID)
		return db.CreateFile(ctx, object)
	})
}
