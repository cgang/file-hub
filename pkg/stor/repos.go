package stor

import (
	"context"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
)

func CreateHomeRepo(ctx context.Context, user *model.User, rootDir string) error {
	repo := &model.Repository{
		Name: user.Username,
		Root: rootDir,
	}

	if err := db.CreateRepository(ctx, repo); err != nil {
		return err
	}

	rootFile := &model.FileObject{
		OwnerID: user.ID,
		RepoID:  repo.ID,
		Name:    "/",
		Path:    "/",
		IsDir:   true,
	}

	if err := db.CreateFile(ctx, rootFile); err != nil {
		return err
	}

	return nil
}

func GetRepository(ctx context.Context, name string) (*model.Repository, error) {
	// TODO add caching layer here
	return db.GetRepositoryByName(ctx, name)
}

// ForUser creates a Storage instance for the given user based on their HomeDir
func GetHomeRepo(ctx context.Context, user *model.User) (*model.Repository, error) {
	return db.GetRepositoryByName(ctx, user.Username)
}
