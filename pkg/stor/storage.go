package stor

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/cgang/file-hub/pkg/config"
	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
)

// Global config variable
var (
	globalConfig *config.Config
	configOnce   sync.Once
)

// InitConfig initializes the global config
func InitConfig(cfg *config.Config) {
	configOnce.Do(func() {
		globalConfig = cfg
	})
}

func GetFileInfo(ctx context.Context, resource *model.Resource) (*model.FileObject, error) {
	return db.GetFile(ctx, resource.ReposID, resource.Path)
}

func ListDir(ctx context.Context, parent *model.FileObject) ([]*model.FileObject, error) {
	return db.GetChildFiles(ctx, parent.ID)
}

func WriteToFile(ctx context.Context, resource *model.Resource, dataReader io.Reader) error {
	return errors.New("not implemented")
}

func DeleteFile(ctx context.Context, resource *model.Resource) error {
	return errors.New("not implemented")
}

func CreateDir(ctx context.Context, resource *model.Resource) error {
	return errors.New("not implemented")
}

func CopyFile(ctx context.Context, srcResource *model.Resource, destResource *model.Resource) error {
	return errors.New("not implemented")
}

func MoveFile(ctx context.Context, srcResource *model.Resource, destResource *model.Resource) error {
	return errors.New("not implemented")
}

func OpenFile(ctx context.Context, resource *model.Resource) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}
