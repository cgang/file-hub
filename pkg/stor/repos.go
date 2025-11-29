package stor

import (
	"context"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
)

func GetRepository(ctx context.Context, name string) (*model.Repository, error) {
	// TODO add caching layer here
	return db.GetRepositoryByName(ctx, name)
}

// ForUser creates a Storage instance for the given user based on their HomeDir
func GetHomeRepo(ctx context.Context, user *model.User) (*model.Repository, error) {
	return db.GetRepositoryByName(ctx, user.Username)
}
