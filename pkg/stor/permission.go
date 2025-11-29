package stor

import (
	"context"
	"errors"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
)

type Permission int

const (
	PermissionUnknown Permission = iota
	PermissionRead
	PermissionWrite
	PermissionDelete
)

func CheckPermission(ctx context.Context, userID int, resource *model.Resource, perm Permission) error {
	if userID == resource.OwnerID {
		return nil // Owner has all permissions
	}

	share, err := db.GetShareForObject(ctx, userID, resource)
	if err != nil {
		return err
	}

	if share == nil {
		return errors.New("object not shared with user")
	}

	if perm == PermissionRead {
		return nil // Read permission granted
	}

	// TODO handle write and delete permissions based on share settings
	return errors.New("permission denied")
}
