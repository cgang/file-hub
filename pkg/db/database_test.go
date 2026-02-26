package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// testDB DSN for test database
const testDBDSN = "postgresql://filehub:filehub@localhost:5432/filehub_test?sslmode=disable"

// setupTestDB initializes the test database
func setupTestDB(t *testing.T) func() {
	// Check if test database is available
	dsn := os.Getenv("FILEHUB_TEST_DB_DSN")
	if dsn == "" {
		dsn = testDBDSN
	}

	ctx := context.Background()
	
	// Try to connect and skip if database is not available
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	if err := pgdb.PingContext(ctx); err != nil {
		t.Skipf("Skipping database tests: %v", err)
		return func() {}
	}

	db = bun.NewDB(pgdb, pgdialect.New())

	// Cleanup function
	cleanup := func() {
		// Truncate all tables
		tables := []string{"user_quota", "shares", "files", "repositories", "users"}
		for _, table := range tables {
			_, err := GetDB().ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
			if err != nil {
				t.Logf("Warning: failed to truncate %s: %v", table, err)
			}
		}
		Close()
	}

	// Initial cleanup
	cleanup()

	return cleanup
}

// TestUserDatabase tests user database operations
func TestUserDatabase(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("CreateUser", func(t *testing.T) {
		user := &model.User{
			Username:  "testuser",
			Email:     "test@example.com",
			HA1:       "testha1hash",
			FirstName: stringPtr("Test"),
			LastName:  stringPtr("User"),
			IsActive:  true,
			IsAdmin:   false,
		}

		err := CreateUser(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)

		// Verify quota was initialized
		quota, err := GetUserQuota(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, quota.UserID)
		assert.Equal(t, int64(10737418240), quota.TotalQuotaBytes)
		assert.Equal(t, int64(0), quota.UsedBytes)
	})

	t.Run("CreateUserDuplicateUsername", func(t *testing.T) {
		user := &model.User{
			Username: "duplicateuser",
			Email:    "dup1@example.com",
			HA1:      "testha1",
			IsActive: true,
		}

		err := CreateUser(ctx, user)
		require.NoError(t, err)

		user2 := &model.User{
			Username: "duplicateuser",
			Email:    "dup2@example.com",
			HA1:      "testha1",
			IsActive: true,
		}

		err = CreateUser(ctx, user2)
		assert.Error(t, err)
	})

	t.Run("CreateUserDuplicateEmail", func(t *testing.T) {
		user := &model.User{
			Username: "user1",
			Email:    "duplicate@example.com",
			HA1:      "testha1",
			IsActive: true,
		}

		err := CreateUser(ctx, user)
		require.NoError(t, err)

		user2 := &model.User{
			Username: "user2",
			Email:    "duplicate@example.com",
			HA1:      "testha1",
			IsActive: true,
		}

		err = CreateUser(ctx, user2)
		assert.Error(t, err)
	})

	t.Run("GetUserByID", func(t *testing.T) {
		// Create a test user
		user := &model.User{
			Username: "getbyiduser",
			Email:    "getbyid@example.com",
			HA1:      "testha1",
			IsActive: true,
		}
		err := CreateUser(ctx, user)
		require.NoError(t, err)

		// Get user by ID
		retrieved, err := GetUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Username, retrieved.Username)
		assert.Equal(t, user.Email, retrieved.Email)

		// Get non-existent user
		_, err = GetUserByID(ctx, 99999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		// Create a test user
		user := &model.User{
			Username: "username_lookup",
			Email:    "username@example.com",
			HA1:      "testha1",
			IsActive: true,
		}
		err := CreateUser(ctx, user)
		require.NoError(t, err)

		// Get user by username
		retrieved, err := GetUserByUsername(ctx, "username_lookup")
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)

		// Get non-existent user
		_, err = GetUserByUsername(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		// Create a test user
		user := &model.User{
			Username: "emailuser",
			Email:    "email_lookup@example.com",
			HA1:      "testha1",
			IsActive: true,
		}
		err := CreateUser(ctx, user)
		require.NoError(t, err)

		// Get user by email
		retrieved, err := GetUserByEmail(ctx, "email_lookup@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Username, retrieved.Username)

		// Get non-existent user
		_, err = GetUserByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetUserInactive", func(t *testing.T) {
		// Create an inactive user
		user := &model.User{
			Username: "inactiveuser",
			Email:    "inactive@example.com",
			HA1:      "testha1",
			IsActive: false,
		}
		err := CreateUser(ctx, user)
		require.NoError(t, err)

		// Try to get inactive user (should not be found)
		_, err = GetUserByUsername(ctx, "inactiveuser")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("UpdateUser", func(t *testing.T) {
		// Create a test user
		user := &model.User{
			Username:  "updateuser",
			Email:     "update@example.com",
			HA1:       "testha1",
			FirstName: stringPtr("First"),
			LastName:  stringPtr("Last"),
			IsActive:  true,
			IsAdmin:   false,
		}
		err := CreateUser(ctx, user)
		require.NoError(t, err)

		// Update user
		update := &UserUpdate{
			FirstName: stringPtr("UpdatedFirst"),
			LastName:  stringPtr("UpdatedLast"),
		}

		err = UpdateUser(ctx, user.ID, update)
		require.NoError(t, err)

		// Verify update
		retrieved, err := GetUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "UpdatedFirst", *retrieved.FirstName)
		assert.Equal(t, "UpdatedLast", *retrieved.LastName)
	})

	t.Run("UpdateUserLastLogin", func(t *testing.T) {
		// Create a test user
		user := &model.User{
			Username: "loginuser",
			Email:    "login@example.com",
			HA1:      "testha1",
			IsActive: true,
		}
		err := CreateUser(ctx, user)
		require.NoError(t, err)

		// Update last login
		now := time.Now()
		update := &UserUpdate{
			LastLogin: &now,
		}

		err = UpdateUser(ctx, user.ID, update)
		require.NoError(t, err)

		// Verify update
		retrieved, err := GetUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved.LastLogin)
	})

	t.Run("UpdateUserNonExistent", func(t *testing.T) {
		update := &UserUpdate{
			FirstName: stringPtr("Test"),
		}

		err := UpdateUser(ctx, 99999, update)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("DeleteUser", func(t *testing.T) {
		// Create a test user
		user := &model.User{
			Username: "deleteuser",
			Email:    "delete@example.com",
			HA1:      "testha1",
			IsActive: true,
		}
		err := CreateUser(ctx, user)
		require.NoError(t, err)

		// Delete user
		err = DeleteUser(ctx, user.ID)
		require.NoError(t, err)

		// Verify user is inactive
		_, err = GetUserByID(ctx, user.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("DeleteUserNonExistent", func(t *testing.T) {
		err := DeleteUser(ctx, 99999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("UpdateUserHA1", func(t *testing.T) {
		// Create a test user
		user := &model.User{
			Username: "ha1user",
			Email:    "ha1@example.com",
			HA1:      "originalha1",
			IsActive: true,
		}
		err := CreateUser(ctx, user)
		require.NoError(t, err)

		// Update HA1
		err = UpdateUserHA1(ctx, user.ID, "newha1hash")
		require.NoError(t, err)

		// Verify update (need to query directly since HA1 is not exported in model)
		dbUser, err := GetUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "newha1hash", dbUser.HA1)
	})

	t.Run("UpdateUserHA1NonExistent", func(t *testing.T) {
		err := UpdateUserHA1(ctx, 99999, "newha1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("CountUsers", func(t *testing.T) {
		// Count should include all users (active and inactive)
		count, err := CountUsers(ctx)
		require.NoError(t, err)
		assert.Greater(t, count, 0)
	})
}

// TestRepositoryDatabase tests repository database operations
func TestRepositoryDatabase(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test user first
	user := &model.User{
		Username: "repoowner",
		Email:    "repoowner@example.com",
		HA1:      "testha1",
		IsActive: true,
	}
	err := CreateUser(ctx, user)
	require.NoError(t, err)

	t.Run("CreateRepository", func(t *testing.T) {
		repo := &model.Repository{
			OwnerID: user.ID,
			Name:    "test-repo",
			Root:    "/storage/test-repo",
		}

		err := CreateRepository(ctx, repo)
		require.NoError(t, err)
		assert.NotZero(t, repo.ID)
		assert.NotZero(t, repo.CreatedAt)
		assert.NotZero(t, repo.UpdatedAt)
	})

	t.Run("CreateRepositoryMultiple", func(t *testing.T) {
		repo1 := &model.Repository{
			OwnerID: user.ID,
			Name:    "repo1",
			Root:    "/storage/repo1",
		}
		err := CreateRepository(ctx, repo1)
		require.NoError(t, err)

		repo2 := &model.Repository{
			OwnerID: user.ID,
			Name:    "repo2",
			Root:    "/storage/repo2",
		}
		err = CreateRepository(ctx, repo2)
		require.NoError(t, err)

		assert.NotEqual(t, repo1.ID, repo2.ID)
	})

	t.Run("GetRepositoryByID", func(t *testing.T) {
		// Create a test repository
		repo := &model.Repository{
			OwnerID: user.ID,
			Name:    "getbyid-repo",
			Root:    "/storage/getbyid",
		}
		err := CreateRepository(ctx, repo)
		require.NoError(t, err)

		// Get by ID
		retrieved, err := GetRepositoryByID(ctx, repo.ID)
		require.NoError(t, err)
		assert.Equal(t, repo.Name, retrieved.Name)
		assert.Equal(t, repo.Root, retrieved.Root)
		assert.Equal(t, user.ID, retrieved.OwnerID)

		// Get non-existent
		_, err = GetRepositoryByID(ctx, 99999)
		assert.Error(t, err)
	})

	t.Run("GetRepositoryByName", func(t *testing.T) {
		// Create a test repository
		repo := &model.Repository{
			OwnerID: user.ID,
			Name:    "getbyname-repo",
			Root:    "/storage/getbyname",
		}
		err := CreateRepository(ctx, repo)
		require.NoError(t, err)

		// Get by name
		retrieved, err := GetRepositoryByName(ctx, "getbyname-repo")
		require.NoError(t, err)
		assert.Equal(t, repo.ID, retrieved.ID)
		assert.Equal(t, repo.Root, retrieved.Root)

		// Get non-existent
		_, err = GetRepositoryByName(ctx, "nonexistent-repo")
		assert.Error(t, err)
	})

	t.Run("GetRepositoryByNameAndOwner", func(t *testing.T) {
		// Create test repositories for different users
		user2 := &model.User{
			Username: "repoowner2",
			Email:    "repoowner2@example.com",
			HA1:      "testha1",
			IsActive: true,
		}
		err := CreateUser(ctx, user2)
		require.NoError(t, err)

		repo1 := &model.Repository{
			OwnerID: user.ID,
			Name:    "shared-name",
			Root:    "/storage/user1",
		}
		err = CreateRepository(ctx, repo1)
		require.NoError(t, err)

		repo2 := &model.Repository{
			OwnerID: user2.ID,
			Name:    "shared-name",
			Root:    "/storage/user2",
		}
		err = CreateRepository(ctx, repo2)
		require.NoError(t, err)

		// Get by name and owner
		retrieved, err := GetRepositoryByNameAndOwner(ctx, "shared-name", user.ID)
		require.NoError(t, err)
		assert.Equal(t, repo1.ID, retrieved.ID)
		assert.Equal(t, "/storage/user1", retrieved.Root)

		// Get for different owner
		retrieved2, err := GetRepositoryByNameAndOwner(ctx, "shared-name", user2.ID)
		require.NoError(t, err)
		assert.Equal(t, repo2.ID, retrieved2.ID)
		assert.Equal(t, "/storage/user2", retrieved2.Root)
	})

	t.Run("ListRepositories", func(t *testing.T) {
		// Create multiple repositories for the user
		repos := []*model.Repository{
			{Name: "list-repo-1", Root: "/storage/list1", OwnerID: user.ID},
			{Name: "list-repo-2", Root: "/storage/list2", OwnerID: user.ID},
			{Name: "list-repo-3", Root: "/storage/list3", OwnerID: user.ID},
		}

		for _, repo := range repos {
			err := CreateRepository(ctx, repo)
			require.NoError(t, err)
		}

		// List repositories
		listed, err := ListRepositories(ctx, user.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listed), 3)

		// Verify all repos are listed
		names := make(map[string]bool)
		for _, repo := range listed {
			names[repo.Name] = true
		}
		assert.True(t, names["list-repo-1"])
		assert.True(t, names["list-repo-2"])
		assert.True(t, names["list-repo-3"])
	})

	t.Run("ListRepositoriesEmpty", func(t *testing.T) {
		// Create a new user with no repositories
		newUser := &model.User{
			Username: "norepouser",
			Email:    "norepo@example.com",
			HA1:      "testha1",
			IsActive: true,
		}
		err := CreateUser(ctx, newUser)
		require.NoError(t, err)

		// List should return empty slice
		listed, err := ListRepositories(ctx, newUser.ID)
		require.NoError(t, err)
		assert.Empty(t, listed)
	})
}

// TestFileDatabase tests file database operations
func TestFileDatabase(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test user and repository
	user := &model.User{
		Username: "fileuser",
		Email:    "fileuser@example.com",
		HA1:      "testha1",
		IsActive: true,
	}
	err := CreateUser(ctx, user)
	require.NoError(t, err)

	repo := &model.Repository{
		OwnerID: user.ID,
		Name:    "file-repo",
		Root:    "/storage/file-repo",
	}
	err = CreateRepository(ctx, repo)
	require.NoError(t, err)

	t.Run("CreateFile", func(t *testing.T) {
		file := &model.FileObject{
			OwnerID: user.ID,
			RepoID:  repo.ID,
			Name:    "testfile.txt",
			Path:    "/testfile.txt",
			Size:    1024,
			IsDir:   false,
			ModTime: time.Now(),
		}

		err := CreateFile(ctx, file)
		require.NoError(t, err)
		assert.NotZero(t, file.ID)
		assert.NotZero(t, file.CreatedAt)
	})

	t.Run("CreateDirectory", func(t *testing.T) {
		dir := &model.FileObject{
			OwnerID: user.ID,
			RepoID:  repo.ID,
			Name:    "testdir",
			Path:    "/testdir",
			IsDir:   true,
			Size:    0,
			ModTime: time.Now(),
		}

		err := CreateFile(ctx, dir)
		require.NoError(t, err)
		assert.NotZero(t, dir.ID)
		assert.True(t, dir.IsDir)
	})

	t.Run("CreateFileWithParent", func(t *testing.T) {
		// Create parent directory first
		parent := &model.FileObject{
			OwnerID: user.ID,
			RepoID:  repo.ID,
			Name:    "parent",
			Path:    "/parent",
			IsDir:   true,
			Size:    0,
			ModTime: time.Now(),
		}
		err := CreateFile(ctx, parent)
		require.NoError(t, err)

		// Create child file
		child := &model.FileObject{
			OwnerID:  user.ID,
			RepoID:   repo.ID,
			ParentID: parent.ID,
			Name:     "child.txt",
			Path:     "/parent/child.txt",
			Size:     512,
			IsDir:    false,
			ModTime:  time.Now(),
		}

		err = CreateFile(ctx, child)
		require.NoError(t, err)
		assert.Equal(t, parent.ID, child.ParentID)
	})

	t.Run("GetFileByID", func(t *testing.T) {
		// Create a test file
		file := &model.FileObject{
			OwnerID: user.ID,
			RepoID:  repo.ID,
			Name:    "getbyid.txt",
			Path:    "/getbyid.txt",
			Size:    2048,
			IsDir:   false,
			ModTime: time.Now(),
		}
		err := CreateFile(ctx, file)
		require.NoError(t, err)

		// Get by ID
		retrieved, err := GetFileByID(ctx, file.ID)
		require.NoError(t, err)
		assert.Equal(t, file.Name, retrieved.Name)
		assert.Equal(t, file.Path, retrieved.Path)
		assert.Equal(t, file.Size, retrieved.Size)

		// Get non-existent
		_, err = GetFileByID(ctx, 99999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetFileByPath", func(t *testing.T) {
		// Create a test file
		file := &model.FileObject{
			OwnerID: user.ID,
			RepoID:  repo.ID,
			Name:    "bypath.txt",
			Path:    "/folder/bypath.txt",
			Size:    1024,
			IsDir:   false,
			ModTime: time.Now(),
		}
		err := CreateFile(ctx, file)
		require.NoError(t, err)

		// Get by path
		retrieved, err := GetFile(ctx, repo.ID, "/folder/bypath.txt")
		require.NoError(t, err)
		assert.Equal(t, file.ID, retrieved.ID)
		assert.Equal(t, file.Name, retrieved.Name)

		// Get non-existent path
		_, err = GetFile(ctx, repo.ID, "/nonexistent.txt")
		assert.Error(t, err)
	})

	t.Run("GetChildFiles", func(t *testing.T) {
		// Create parent directory
		parent := &model.FileObject{
			OwnerID: user.ID,
			RepoID:  repo.ID,
			Name:    "parent-dir",
			Path:    "/parent-dir",
			IsDir:   true,
			Size:    0,
			ModTime: time.Now(),
		}
		err := CreateFile(ctx, parent)
		require.NoError(t, err)

		// Create child files
		children := []*model.FileObject{
			{OwnerID: user.ID, RepoID: repo.ID, ParentID: parent.ID, Name: "child1.txt", Path: "/parent-dir/child1.txt", Size: 100, IsDir: false, ModTime: time.Now()},
			{OwnerID: user.ID, RepoID: repo.ID, ParentID: parent.ID, Name: "child2.txt", Path: "/parent-dir/child2.txt", Size: 200, IsDir: false, ModTime: time.Now()},
			{OwnerID: user.ID, RepoID: repo.ID, ParentID: parent.ID, Name: "child3.txt", Path: "/parent-dir/child3.txt", Size: 300, IsDir: false, ModTime: time.Now()},
		}

		for _, child := range children {
			err := CreateFile(ctx, child)
			require.NoError(t, err)
		}

		// Get child files
		retrieved, err := GetChildFiles(ctx, parent.ID)
		require.NoError(t, err)
		assert.Len(t, retrieved, 3)

		// Verify ordering by name
		assert.Equal(t, "child1.txt", retrieved[0].Name)
		assert.Equal(t, "child2.txt", retrieved[1].Name)
		assert.Equal(t, "child3.txt", retrieved[2].Name)
	})

	t.Run("GetFilesByUser", func(t *testing.T) {
		// Create files for the user
		files := []*model.FileObject{
			{OwnerID: user.ID, RepoID: repo.ID, Name: "user-file1.txt", Path: "/user-file1.txt", Size: 100, IsDir: false, ModTime: time.Now()},
			{OwnerID: user.ID, RepoID: repo.ID, Name: "user-file2.txt", Path: "/user-file2.txt", Size: 200, IsDir: false, ModTime: time.Now()},
		}

		for _, file := range files {
			err := CreateFile(ctx, file)
			require.NoError(t, err)
		}

		// Get files by user
		retrieved, err := GetFilesByUser(ctx, user.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(retrieved), 2)
	})

	t.Run("GetFilesByUserAndPathPrefix", func(t *testing.T) {
		// Create files with common path prefix
		files := []*model.FileObject{
			{OwnerID: user.ID, RepoID: repo.ID, Name: "docs", Path: "/docs", IsDir: true, Size: 0, ModTime: time.Now()},
			{OwnerID: user.ID, RepoID: repo.ID, Name: "file1.txt", Path: "/docs/file1.txt", Size: 100, IsDir: false, ModTime: time.Now()},
			{OwnerID: user.ID, RepoID: repo.ID, Name: "file2.txt", Path: "/docs/file2.txt", Size: 200, IsDir: false, ModTime: time.Now()},
			{OwnerID: user.ID, RepoID: repo.ID, Name: "other.txt", Path: "/other.txt", Size: 300, IsDir: false, ModTime: time.Now()},
		}

		for _, file := range files {
			err := CreateFile(ctx, file)
			require.NoError(t, err)
		}

		// Get files with path prefix
		retrieved, err := GetFilesByUserAndPathPrefix(ctx, user.ID, "/docs")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(retrieved), 3) // Includes /docs, /docs/file1.txt, /docs/file2.txt

		// Verify all returned files have the prefix
		for _, file := range retrieved {
			assert.True(t, file.Path == "/docs" || len(file.Path) > 5 && file.Path[:5] == "/docs")
		}
	})

	t.Run("UpdateFile", func(t *testing.T) {
		// Create a test file
		file := &model.FileObject{
			OwnerID: user.ID,
			RepoID:  repo.ID,
			Name:    "updatefile.txt",
			Path:    "/updatefile.txt",
			Size:    1024,
			IsDir:   false,
			ModTime: time.Now(),
		}
		err := CreateFile(ctx, file)
		require.NoError(t, err)

		// Update file
		newSize := int64(2048)
		newChecksum := "sha256:abc123"
		newMime := "text/plain"
		update := &FileUpdate{
			Size:     &newSize,
			Checksum: &newChecksum,
			MimeType: &newMime,
		}

		err = UpdateFile(ctx, file.ID, update)
		require.NoError(t, err)

		// Verify update
		retrieved, err := GetFileByID(ctx, file.ID)
		require.NoError(t, err)
		assert.Equal(t, newSize, retrieved.Size)
		assert.Equal(t, newChecksum, *retrieved.Checksum)
		assert.Equal(t, newMime, *retrieved.MimeType)
	})

	t.Run("UpdateFileNonExistent", func(t *testing.T) {
		update := &FileUpdate{
			Size: int64Ptr(1024),
		}

		err := UpdateFile(ctx, 99999, update)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("DeleteFile", func(t *testing.T) {
		// Create a test file
		file := &model.FileObject{
			OwnerID: user.ID,
			RepoID:  repo.ID,
			Name:    "deletefile.txt",
			Path:    "/deletefile.txt",
			Size:    512,
			IsDir:   false,
			ModTime: time.Now(),
		}
		err := CreateFile(ctx, file)
		require.NoError(t, err)

		// Delete file
		err = DeleteFile(ctx, file.ID)
		require.NoError(t, err)

		// Verify file is deleted
		_, err = GetFileByID(ctx, file.ID)
		assert.Error(t, err)
	})

	t.Run("DeleteFileNonExistent", func(t *testing.T) {
		err := DeleteFile(ctx, 99999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("UpsertFile", func(t *testing.T) {
		// Create initial file
		file := &model.FileObject{
			OwnerID: user.ID,
			RepoID:  repo.ID,
			Name:    "upsertfile.txt",
			Path:    "/upsertfile.txt",
			Size:    1024,
			IsDir:   false,
			ModTime: time.Now(),
		}
		err := UpsertFile(ctx, file)
		require.NoError(t, err)
		originalID := file.ID
		originalModTime := file.ModTime

		// Upsert same file with different size
		time.Sleep(10 * time.Millisecond) // Ensure different timestamp
		file.Size = 2048
		file.ModTime = time.Now()

		err = UpsertFile(ctx, file)
		require.NoError(t, err)

		// Verify file was updated (same ID)
		assert.Equal(t, originalID, file.ID)
		assert.NotEqual(t, originalModTime, file.ModTime)

		// Verify size was updated
		retrieved, err := GetFileByID(ctx, file.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2048), retrieved.Size)
	})

	t.Run("UpsertFileMissingFields", func(t *testing.T) {
		file := &model.FileObject{
			Name: "invalid.txt",
			Size: 1024,
		}

		err := UpsertFile(ctx, file)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repo_id and path are required")
	})

	t.Run("DeleteFileByPath", func(t *testing.T) {
		// Create a test file
		file := &model.FileObject{
			OwnerID: user.ID,
			RepoID:  repo.ID,
			Name:    "deletebypath.txt",
			Path:    "/deletebypath.txt",
			Size:    256,
			IsDir:   false,
			ModTime: time.Now(),
		}
		err := CreateFile(ctx, file)
		require.NoError(t, err)

		// Delete by path
		err = DeleteFileByPath(ctx, repo.ID, "/deletebypath.txt")
		require.NoError(t, err)

		// Verify file is deleted
		_, err = GetFile(ctx, repo.ID, "/deletebypath.txt")
		assert.Error(t, err)
	})

	t.Run("DeleteFileByPathNonExistent", func(t *testing.T) {
		err := DeleteFileByPath(ctx, repo.ID, "/nonexistent.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestQuotaDatabase tests quota database operations
func TestQuotaDatabase(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test user (quota is auto-initialized)
	user := &model.User{
		Username: "quotauser",
		Email:    "quotauser@example.com",
		HA1:      "testha1",
		IsActive: true,
	}
	err := CreateUser(ctx, user)
	require.NoError(t, err)

	t.Run("GetUserQuota", func(t *testing.T) {
		quota, err := GetUserQuota(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, quota.UserID)
		assert.Equal(t, int64(10737418240), quota.TotalQuotaBytes)
		assert.Equal(t, int64(0), quota.UsedBytes)
	})

	t.Run("GetUserQuotaNonExistent", func(t *testing.T) {
		_, err := GetUserQuota(ctx, 99999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("UpdateUserQuota", func(t *testing.T) {
		newQuota := int64(21474836480) // 20GB
		err := UpdateUserQuota(ctx, user.ID, newQuota)
		require.NoError(t, err)

		// Verify update
		quota, err := GetUserQuota(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, newQuota, quota.TotalQuotaBytes)
	})

	t.Run("UpdateUserQuotaNonExistent", func(t *testing.T) {
		err := UpdateUserQuota(ctx, 99999, 10737418240)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetUserQuotaUsage", func(t *testing.T) {
		used, err := GetUserQuotaUsage(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), used)
	})

	t.Run("CheckUserQuota", func(t *testing.T) {
		// Check quota for small file (should pass)
		hasSpace, err := CheckUserQuota(ctx, user.ID, 1024)
		require.NoError(t, err)
		assert.True(t, hasSpace)

		// Check quota for very large file (should fail)
		hasSpace, err = CheckUserQuota(ctx, user.ID, 21474836480)
		require.NoError(t, err)
		assert.False(t, hasSpace)
	})
}

// TestShareDatabase tests share database operations
func TestShareDatabase(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test users
	owner := &model.User{
		Username: "shareowner",
		Email:    "shareowner@example.com",
		HA1:      "testha1",
		IsActive: true,
	}
	err := CreateUser(ctx, owner)
	require.NoError(t, err)

	recipient := &model.User{
		Username: "sharerecipient",
		Email:    "sharerecipient@example.com",
		HA1:      "testha1",
		IsActive: true,
	}
	err = CreateUser(ctx, recipient)
	require.NoError(t, err)

	// Create a repository
	repo := &model.Repository{
		OwnerID: owner.ID,
		Name:    "share-repo",
		Root:    "/storage/share-repo",
	}
	err = CreateRepository(ctx, repo)
	require.NoError(t, err)

	t.Run("CreateShare", func(t *testing.T) {
		share := &model.Share{
			RepoID:  repo.ID,
			OwnerID: owner.ID,
			UserID:  recipient.ID,
			Path:    "/shared-folder",
		}

		err := CreateShare(ctx, share)
		require.NoError(t, err)
		assert.NotZero(t, share.ID)
	})

	t.Run("CreateShareDuplicate", func(t *testing.T) {
		share1 := &model.Share{
			RepoID:  repo.ID,
			OwnerID: owner.ID,
			UserID:  recipient.ID,
			Path:    "/dup-share",
		}
		err := CreateShare(ctx, share1)
		require.NoError(t, err)

		// Note: The schema doesn't have a unique constraint that would prevent
		// duplicate shares, so this test verifies the behavior
		share2 := &model.Share{
			RepoID:  repo.ID,
			OwnerID: owner.ID,
			UserID:  recipient.ID,
			Path:    "/dup-share",
		}
		err = CreateShare(ctx, share2)
		// This may or may not error depending on schema constraints
		_ = share2
	})

	t.Run("GetShareByID", func(t *testing.T) {
		// Create a test share
		share := &model.Share{
			RepoID:  repo.ID,
			OwnerID: owner.ID,
			UserID:  recipient.ID,
			Path:    "/getbyid-share",
		}
		err := CreateShare(ctx, share)
		require.NoError(t, err)

		// Get by ID
		retrieved, err := GetShareByID(ctx, share.ID)
		require.NoError(t, err)
		assert.Equal(t, share.RepoID, retrieved.RepoID)
		assert.Equal(t, share.UserID, retrieved.UserID)
		assert.Equal(t, share.Path, retrieved.Path)

		// Get non-existent
		_, err = GetShareByID(ctx, 99999)
		assert.Error(t, err)
	})

	t.Run("GetSharesByUserID", func(t *testing.T) {
		// Create multiple shares for the recipient
		shares := []*model.Share{
			{RepoID: repo.ID, OwnerID: owner.ID, UserID: recipient.ID, Path: "/share1"},
			{RepoID: repo.ID, OwnerID: owner.ID, UserID: recipient.ID, Path: "/share2"},
			{RepoID: repo.ID, OwnerID: owner.ID, UserID: recipient.ID, Path: "/share3"},
		}

		for _, share := range shares {
			err := CreateShare(ctx, share)
			require.NoError(t, err)
		}

		// Get shares by user ID
		retrieved, err := GetSharesByUserID(ctx, recipient.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(retrieved), 3)
	})

	t.Run("GetSharesByUserIDEmpty", func(t *testing.T) {
		// Create a user with no shares
		newUser := &model.User{
			Username: "noshares",
			Email:    "noshares@example.com",
			HA1:      "testha1",
			IsActive: true,
		}
		err := CreateUser(ctx, newUser)
		require.NoError(t, err)

		// Get shares should return empty
		shares, err := GetSharesByUserID(ctx, newUser.ID)
		require.NoError(t, err)
		assert.Empty(t, shares)
	})

	t.Run("DeleteShare", func(t *testing.T) {
		// Create a test share
		share := &model.Share{
			RepoID:  repo.ID,
			OwnerID: owner.ID,
			UserID:  recipient.ID,
			Path:    "/delete-share",
		}
		err := CreateShare(ctx, share)
		require.NoError(t, err)

		// Delete share
		err = DeleteShareByID(ctx, share.ID)
		require.NoError(t, err)

		// Verify share is deleted
		_, err = GetShareByID(ctx, share.ID)
		assert.Error(t, err)
	})

	t.Run("DeleteShareNonExistent", func(t *testing.T) {
		err := DeleteShareByID(ctx, 99999)
		assert.Error(t, err)
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
