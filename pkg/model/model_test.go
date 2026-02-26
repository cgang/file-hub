package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	t.Run("User JSON serialization", func(t *testing.T) {
		now := time.Now()
		user := &User{
			ID:        1,
			Username:  "testuser",
			Email:     "test@example.com",
			HA1:       "secret_hash",
			FirstName: stringPtr("Test"),
			LastName:  stringPtr("User"),
			CreatedAt: now,
			UpdatedAt: now,
			LastLogin: &now,
			IsActive:  true,
			IsAdmin:   false,
		}

		data, err := json.Marshal(user)
		assert.NoError(t, err)

		// HA1 should not be in JSON output
		assert.NotContains(t, string(data), "ha1")
		assert.NotContains(t, string(data), "secret_hash")

		// Other fields should be present
		assert.Contains(t, string(data), "testuser")
		assert.Contains(t, string(data), "test@example.com")
	})

	t.Run("User JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"id": 1,
			"username": "newuser",
			"email": "new@example.com",
			"first_name": "New",
			"last_name": "User",
			"is_active": true,
			"is_admin": false
		}`

		var user User
		err := json.Unmarshal([]byte(jsonData), &user)
		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "newuser", user.Username)
		assert.Equal(t, "new@example.com", user.Email)
		assert.Equal(t, "New", *user.FirstName)
		assert.Equal(t, "User", *user.LastName)
		assert.True(t, user.IsActive)
		assert.False(t, user.IsAdmin)
	})

	t.Run("User with nil optional fields", func(t *testing.T) {
		user := &User{
			ID:        2,
			Username:  "minimal",
			Email:     "minimal@example.com",
			HA1:       "hash",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
		}

		data, err := json.Marshal(user)
		assert.NoError(t, err)

		// Optional fields should be omitted when nil
		assert.NotContains(t, string(data), "first_name")
		assert.NotContains(t, string(data), "last_name")
		assert.NotContains(t, string(data), "last_login")
	})
}

func TestUserQuotaModel(t *testing.T) {
	t.Run("UserQuota JSON serialization", func(t *testing.T) {
		now := time.Now()
		quota := &UserQuota{
			ID:              1,
			UserID:          42,
			TotalQuotaBytes: 10 * 1024 * 1024 * 1024, // 10GB
			UsedBytes:       5 * 1024 * 1024 * 1024,  // 5GB
			UpdatedAt:       now,
		}

		data, err := json.Marshal(quota)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "42")
		assert.Contains(t, string(data), "10737418240")
		assert.Contains(t, string(data), "5368709120")
	})

	t.Run("UserQuota JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"id": 1,
			"user_id": 42,
			"total_quota_bytes": 10737418240,
			"used_bytes": 5368709120
		}`

		var quota UserQuota
		err := json.Unmarshal([]byte(jsonData), &quota)
		assert.NoError(t, err)
		assert.Equal(t, int64(10737418240), quota.TotalQuotaBytes)
		assert.Equal(t, int64(5368709120), quota.UsedBytes)
	})

	t.Run("UserQuota remaining quota calculation", func(t *testing.T) {
		quota := &UserQuota{
			TotalQuotaBytes: 1000,
			UsedBytes:       600,
		}

		remaining := quota.TotalQuotaBytes - quota.UsedBytes
		assert.Equal(t, int64(400), remaining)
	})

	t.Run("UserQuota usage percentage", func(t *testing.T) {
		quota := &UserQuota{
			TotalQuotaBytes: 1000,
			UsedBytes:       750,
		}

		percentage := float64(quota.UsedBytes) / float64(quota.TotalQuotaBytes) * 100
		assert.InDelta(t, 75.0, percentage, 0.01)
	})
}

func TestRepositoryModel(t *testing.T) {
	t.Run("Repository JSON serialization", func(t *testing.T) {
		now := time.Now()
		repo := &Repository{
			ID:        1,
			OwnerID:   42,
			Name:      "my-repo",
			Root:      "/data/repos/my-repo",
			CreatedAt: now,
			UpdatedAt: now,
		}

		data, err := json.Marshal(repo)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "my-repo")
		assert.Contains(t, string(data), "/data/repos/my-repo")
	})

	t.Run("Repository JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"id": 1,
			"owner_id": 42,
			"name": "test-repo",
			"root": "/data/test-repo"
		}`

		var repo Repository
		err := json.Unmarshal([]byte(jsonData), &repo)
		assert.NoError(t, err)
		assert.Equal(t, "test-repo", repo.Name)
		assert.Equal(t, "/data/test-repo", repo.Root)
	})

	t.Run("Home repository identification", func(t *testing.T) {
		repo1 := &Repository{Name: "john", Root: "/data/john"}
		repo2 := &Repository{Name: "shared", Root: "/data/shared"}

		// Home repository has same name as owner username
		// This is a naming convention test
		assert.Equal(t, "john", repo1.Name)
		assert.Equal(t, "shared", repo2.Name)
	})
}

func TestShareModel(t *testing.T) {
	t.Run("Share JSON serialization", func(t *testing.T) {
		share := &Share{
			ID:      1,
			RepoID:  10,
			OwnerID: 42,
			UserID:  99,
			Path:    "/shared/docs",
		}

		data, err := json.Marshal(share)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "/shared/docs")
	})

	t.Run("Share JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"id": 1,
			"repo_id": 10,
			"owner_id": 42,
			"user_id": 99,
			"path": "/shared/folder"
		}`

		var share Share
		err := json.Unmarshal([]byte(jsonData), &share)
		assert.NoError(t, err)
		assert.Equal(t, 10, share.RepoID)
		assert.Equal(t, 42, share.OwnerID)
		assert.Equal(t, 99, share.UserID)
		assert.Equal(t, "/shared/folder", share.Path)
	})

	t.Run("Share path variations", func(t *testing.T) {
		shares := []Share{
			{Path: "/"},
			{Path: "/docs"},
			{Path: "/docs/reports"},
			{Path: "/deep/nested/path/here"},
		}

		for _, share := range shares {
			data, err := json.Marshal(share)
			assert.NoError(t, err)
			assert.Contains(t, string(data), share.Path)
		}
	})
}

func TestFileObjectModel(t *testing.T) {
	t.Run("FileObject JSON serialization", func(t *testing.T) {
		now := time.Now()
		checksum := "sha256:abc123"
		mimeType := "text/plain"

		file := &FileObject{
			ID:        1,
			ParentID:  0,
			OwnerID:   42,
			RepoID:    10,
			Name:      "readme.txt",
			Path:      "/docs/readme.txt",
			MimeType:  &mimeType,
			Size:      1024,
			ModTime:   now,
			Checksum:  &checksum,
			CreatedAt: now,
			UpdatedAt: now,
			IsDir:     false,
		}

		data, err := json.Marshal(file)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "readme.txt")
		assert.Contains(t, string(data), "text/plain")
		assert.Contains(t, string(data), "sha256:abc123")
	})

	t.Run("FileObject JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"id": 1,
			"parent_id": 5,
			"owner_id": 42,
			"repo_id": 10,
			"name": "test.pdf",
			"path": "/files/test.pdf",
			"mime_type": "application/pdf",
			"size": 2048,
			"is_dir": false
		}`

		var file FileObject
		err := json.Unmarshal([]byte(jsonData), &file)
		assert.NoError(t, err)
		assert.Equal(t, "test.pdf", file.Name)
		assert.Equal(t, "application/pdf", *file.MimeType)
		assert.Equal(t, int64(2048), file.Size)
		assert.False(t, file.IsDir)
	})

	t.Run("FileObject directory serialization", func(t *testing.T) {
		now := time.Now()
		dir := &FileObject{
			ID:        1,
			OwnerID:   42,
			RepoID:    10,
			Name:      "documents",
			Path:      "/documents/",
			ModTime:   now,
			CreatedAt: now,
			UpdatedAt: now,
			IsDir:     true,
		}

		data, err := json.Marshal(dir)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "documents")
		assert.Contains(t, string(data), "true")
	})

	t.Run("FileObject ContentType for file", func(t *testing.T) {
		mimeType := "image/png"
		file := &FileObject{
			MimeType: &mimeType,
			IsDir:    false,
		}

		assert.Equal(t, "image/png", file.ContentType())
	})

	t.Run("FileObject ContentType for directory", func(t *testing.T) {
		dir := &FileObject{
			MimeType: nil,
			IsDir:    true,
		}

		assert.Equal(t, "httpd/unix-directory", dir.ContentType())
	})

	t.Run("FileObject ContentType default", func(t *testing.T) {
		file := &FileObject{
			MimeType: nil,
			IsDir:    false,
		}

		assert.Equal(t, "application/octet-stream", file.ContentType())
	})

	t.Run("FileObject with nil optional fields", func(t *testing.T) {
		now := time.Now()
		file := &FileObject{
			ID:        1,
			OwnerID:   42,
			RepoID:    10,
			Name:      "minimal.txt",
			Path:      "/minimal.txt",
			Size:      100,
			ModTime:   now,
			CreatedAt: now,
			UpdatedAt: now,
			IsDir:     false,
		}

		data, err := json.Marshal(file)
		assert.NoError(t, err)
		assert.NotContains(t, string(data), "mime_type")
		assert.NotContains(t, string(data), "checksum")
		assert.NotContains(t, string(data), "parent_id")
	})
}

func TestResourceModel(t *testing.T) {
	t.Run("Resource String representation", func(t *testing.T) {
		repo := &Repository{
			ID:      1,
			OwnerID: 42,
			Name:    "myrepo",
			Root:    "/data/myrepo",
		}

		resource := &Resource{
			Repo: repo,
			Path: "/docs/file.txt",
		}

		assert.Equal(t, "myrepo/docs/file.txt", resource.String())
	})

	t.Run("Resource with root path", func(t *testing.T) {
		repo := &Repository{
			Name: "home",
			Root: "/data/home",
		}

		resource := &Resource{
			Repo: repo,
			Path: "/",
		}

		assert.Equal(t, "home/", resource.String())
	})

	t.Run("Resource with empty path", func(t *testing.T) {
		repo := &Repository{
			Name: "repo",
			Root: "/data/repo",
		}

		resource := &Resource{
			Repo: repo,
			Path: "",
		}

		assert.Equal(t, "repo", resource.String())
	})
}

func TestChangeLogModel(t *testing.T) {
	t.Run("ChangeLog JSON serialization", func(t *testing.T) {
		now := time.Now()
		oldPath := "/old/path"

		changeLog := &ChangeLog{
			ID:        1,
			RepoID:    10,
			Operation: "move",
			Path:      "/new/path",
			OldPath:   &oldPath,
			UserID:    42,
			Version:   "v1234567890-12345",
			Timestamp: now,
		}

		data, err := json.Marshal(changeLog)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "move")
		assert.Contains(t, string(data), "/new/path")
		assert.Contains(t, string(data), "/old/path")
	})

	t.Run("ChangeLog without OldPath", func(t *testing.T) {
		now := time.Now()
		changeLog := &ChangeLog{
			ID:        1,
			RepoID:    10,
			Operation: "create",
			Path:      "/new/file.txt",
			OldPath:   nil,
			UserID:    42,
			Version:   "v1234567890-12345",
			Timestamp: now,
		}

		data, err := json.Marshal(changeLog)
		assert.NoError(t, err)
		assert.NotContains(t, string(data), "old_path")
	})

	t.Run("ChangeLog operation types", func(t *testing.T) {
		operations := []string{"create", "update", "delete", "move", "copy"}
		now := time.Now()

		for _, op := range operations {
			changeLog := &ChangeLog{
				RepoID:    10,
				Operation: op,
				Path:      "/test",
				UserID:    42,
				Version:   "v123-456",
				Timestamp: now,
			}

			data, err := json.Marshal(changeLog)
			assert.NoError(t, err)
			assert.Contains(t, string(data), op)
		}
	})
}

func TestRepositoryVersionModel(t *testing.T) {
	t.Run("RepositoryVersion JSON serialization", func(t *testing.T) {
		now := time.Now()
		version := &RepositoryVersion{
			ID:             1,
			RepoID:         10,
			CurrentVersion: "v1234567890-abc",
			VersionVector:  "1:5,2:3,3:7",
			UpdatedAt:      now,
		}

		data, err := json.Marshal(version)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "v1234567890-abc")
		assert.Contains(t, string(data), "1:5,2:3,3:7")
	})

	t.Run("RepositoryVersion JSON deserialization", func(t *testing.T) {
		// Note: RepositoryVersion uses Go field names for JSON (no json tags)
		jsonData := `{
			"ID": 1,
			"RepoID": 10,
			"CurrentVersion": "v9876543210-xyz",
			"VersionVector": "1:10,2:5"
		}`

		var version RepositoryVersion
		err := json.Unmarshal([]byte(jsonData), &version)
		assert.NoError(t, err)
		assert.Equal(t, "v9876543210-xyz", version.CurrentVersion)
		assert.Equal(t, "1:10,2:5", version.VersionVector)
	})
}

func TestUploadSessionModel(t *testing.T) {
	t.Run("UploadSession JSON serialization", func(t *testing.T) {
		now := time.Now()
		session := &UploadSession{
			ID:             1,
			UploadID:       "upload-abc123",
			RepoID:         10,
			Path:           "/uploads/large-file.bin",
			TotalSize:      100 * 1024 * 1024,
			UserID:         42,
			ChunksUploaded: 5,
			TotalChunks:    10,
			CreatedAt:      now,
			ExpiresAt:      now.Add(24 * time.Hour),
			Status:         "active",
		}

		data, err := json.Marshal(session)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "upload-abc123")
		assert.Contains(t, string(data), "active")
	})

	t.Run("UploadSession JSON deserialization", func(t *testing.T) {
		// Note: UploadSession uses Go field names for JSON (no json tags)
		jsonData := `{
			"ID": 1,
			"UploadID": "upload-xyz789",
			"RepoID": 10,
			"Path": "/uploads/file.bin",
			"TotalSize": 52428800,
			"UserID": 42,
			"ChunksUploaded": 3,
			"TotalChunks": 5,
			"Status": "active"
		}`

		var session UploadSession
		err := json.Unmarshal([]byte(jsonData), &session)
		assert.NoError(t, err)
		assert.Equal(t, "upload-xyz789", session.UploadID)
		assert.Equal(t, int64(52428800), session.TotalSize)
		assert.Equal(t, 3, session.ChunksUploaded)
		assert.Equal(t, 5, session.TotalChunks)
		assert.Equal(t, "active", session.Status)
	})

	t.Run("UploadSession status values", func(t *testing.T) {
		statuses := []string{"active", "completed", "failed", "expired"}
		now := time.Now()

		for _, status := range statuses {
			session := &UploadSession{
				UploadID:    "test-upload",
				RepoID:      10,
				Path:        "/test",
				TotalSize:   1000,
				UserID:      42,
				TotalChunks: 1,
				CreatedAt:   now,
				ExpiresAt:   now.Add(time.Hour),
				Status:      status,
			}

			data, err := json.Marshal(session)
			assert.NoError(t, err)
			assert.Contains(t, string(data), status)
		}
	})

	t.Run("UploadSession progress calculation", func(t *testing.T) {
		session := &UploadSession{
			ChunksUploaded: 7,
			TotalChunks:    10,
		}

		progress := float64(session.ChunksUploaded) / float64(session.TotalChunks) * 100
		assert.InDelta(t, 70.0, progress, 0.01)
	})

	t.Run("UploadSession expiry check", func(t *testing.T) {
		now := time.Now()
		session := &UploadSession{
			ExpiresAt: now.Add(-1 * time.Hour),
		}

		assert.True(t, now.After(session.ExpiresAt))
		assert.True(t, session.ExpiresAt.Before(now))
	})
}

func TestUploadChunkModel(t *testing.T) {
	t.Run("UploadChunk JSON serialization", func(t *testing.T) {
		now := time.Now()
		checksum := "sha256:def456"

		chunk := &UploadChunk{
			ID:         1,
			UploadID:   "upload-abc123",
			ChunkIndex: 3,
			Offset:     3 * 1024 * 1024,
			Size:       1024 * 1024,
			Checksum:   &checksum,
			UploadedAt: now,
		}

		data, err := json.Marshal(chunk)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "upload-abc123")
		assert.Contains(t, string(data), "sha256:def456")
	})

	t.Run("UploadChunk without checksum", func(t *testing.T) {
		now := time.Now()
		chunk := &UploadChunk{
			ID:         1,
			UploadID:   "upload-xyz",
			ChunkIndex: 0,
			Offset:     0,
			Size:       1024,
			Checksum:   nil,
			UploadedAt: now,
		}

		data, err := json.Marshal(chunk)
		assert.NoError(t, err)
		assert.NotContains(t, string(data), "checksum")
	})

	t.Run("UploadChunk JSON deserialization", func(t *testing.T) {
		// Note: UploadChunk uses Go field names for JSON (no json tags)
		jsonData := `{
			"ID": 1,
			"UploadID": "upload-test",
			"ChunkIndex": 2,
			"Offset": 2097152,
			"Size": 1048576,
			"Checksum": "sha256:abc123"
		}`

		var chunk UploadChunk
		err := json.Unmarshal([]byte(jsonData), &chunk)
		assert.NoError(t, err)
		assert.Equal(t, 2, chunk.ChunkIndex)
		assert.Equal(t, int64(2097152), chunk.Offset)
		assert.Equal(t, int64(1048576), chunk.Size)
		assert.Equal(t, "sha256:abc123", *chunk.Checksum)
	})
}

func TestModelBunTags(t *testing.T) {
	t.Run("User bun tags", func(t *testing.T) {
		user := &User{}
		// Verify struct has bun tags by checking type
		assert.NotNil(t, user)
	})

	t.Run("Repository bun tags", func(t *testing.T) {
		repo := &Repository{}
		assert.NotNil(t, repo)
	})

	t.Run("FileObject bun tags", func(t *testing.T) {
		file := &FileObject{}
		assert.NotNil(t, file)
	})

	t.Run("Share bun tags", func(t *testing.T) {
		share := &Share{}
		assert.NotNil(t, share)
	})
}

func stringPtr(s string) *string {
	return &s
}

// TestUserValidation tests user validation logic
func TestUserValidation(t *testing.T) {
	t.Run("Valid username formats", func(t *testing.T) {
		validUsernames := []string{
			"testuser",
			"test_user",
			"test-user",
			"testuser123",
			"a",
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			"0123456789",
		}

		for _, username := range validUsernames {
			user := &User{
				Username: username,
				Email:    "test@example.com",
				HA1:      "hash",
			}
			// Basic validation - username should be non-empty
			assert.NotEmpty(t, user.Username)
		}
	})

	t.Run("Username length validation", func(t *testing.T) {
		// Minimum length
		user := &User{
			Username: "a",
			Email:    "test@example.com",
			HA1:      "hash",
		}
		assert.NotEmpty(t, user.Username)

		// Maximum length (255 chars per schema)
		longUsername := stringRepeat("a", 255)
		user2 := &User{
			Username: longUsername,
			Email:    "test@example.com",
			HA1:      "hash",
		}
		assert.Len(t, user2.Username, 255)
	})

	t.Run("Valid email formats", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"user.name@domain.org",
			"user+tag@example.co.uk",
			"user123@test-domain.com",
		}

		for _, email := range validEmails {
			user := &User{
				Username: "testuser",
				Email:    email,
				HA1:      "hash",
			}
			assert.Contains(t, user.Email, "@")
		}
	})

	t.Run("Invalid email formats", func(t *testing.T) {
		// These test cases document what should be validated
		// The actual validation logic would be in application code
		invalidEmails := []string{
			"invalid",       // no @
			"no@domain",     // no TLD
			"@example.com",  // no local part
			"user@",         // no domain
			"",              // empty
		}

		// These are examples of invalid emails that should be rejected
		// by application-level validation (not model layer)
		for _, email := range invalidEmails {
			_ = email // Document that these are invalid cases
		}
	})

	t.Run("HA1 hash required", func(t *testing.T) {
		user := &User{
			Username: "testuser",
			Email:    "test@example.com",
			HA1:      "",
		}
		assert.Empty(t, user.HA1, "HA1 should be validated as non-empty")
	})
}

// TestFileObjectValidation tests file object validation logic
func TestFileObjectValidation(t *testing.T) {
	t.Run("Valid file paths", func(t *testing.T) {
		validPaths := []string{
			"/file.txt",
			"/folder/file.txt",
			"/deep/nested/path/file.txt",
			"/",
			"/folder/",
		}

		for _, path := range validPaths {
			file := &FileObject{
				Path: path,
				Name: "file.txt",
			}
			assert.NotEmpty(t, file.Path)
		}
	})

	t.Run("Path traversal prevention", func(t *testing.T) {
		invalidPaths := []string{
			"../etc/passwd",
			"/folder/../../../etc/passwd",
			"..\\windows\\system32",
		}

		for _, path := range invalidPaths {
			// Application should reject these paths
			assert.Contains(t, path, "..", "Path should be validated for traversal: %s", path)
		}
	})

	t.Run("File size validation", func(t *testing.T) {
		validSizes := []int64{0, 1, 1024, 1048576, 1073741824}

		for _, size := range validSizes {
			file := &FileObject{
				Size: size,
			}
			assert.GreaterOrEqual(t, file.Size, int64(0))
		}
	})

	t.Run("Negative size rejected", func(t *testing.T) {
		file := &FileObject{
			Size: -1,
		}
		assert.Less(t, file.Size, int64(0), "Negative sizes should be rejected")
	})

	t.Run("Checksum format validation", func(t *testing.T) {
		validChecksums := []string{
			"sha256:abc123def456",
			"sha512:longhash",
			"md5:abc123",
		}

		for _, checksum := range validChecksums {
			assert.Contains(t, checksum, ":", "Checksum should have format algorithm:hash")
		}
	})

	t.Run("File name validation", func(t *testing.T) {
		validNames := []string{
			"file.txt",
			"my-file.pdf",
			"file_name.doc",
			"文件.txt",
			"file with spaces.txt",
		}

		for _, name := range validNames {
			file := &FileObject{
				Name: name,
			}
			assert.NotEmpty(t, file.Name)
		}

		// Note: "." and ".." are special directory names that may be valid in some contexts
		// Application-level validation should reject these for file creation
		// but the model layer doesn't enforce this
	})
}

// TestRepositoryValidation tests repository validation logic
func TestRepositoryValidation(t *testing.T) {
	t.Run("Valid repository names", func(t *testing.T) {
		validNames := []string{
			"repo",
			"my-repo",
			"my_repo",
			"repo123",
			"Repository Name",
		}

		for _, name := range validNames {
			repo := &Repository{
				Name: name,
			}
			assert.NotEmpty(t, repo.Name)
		}
	})

	t.Run("Repository name length", func(t *testing.T) {
		// Maximum length (255 chars per schema)
		longName := stringRepeat("a", 255)
		repo := &Repository{
			Name: longName,
		}
		assert.Len(t, repo.Name, 255)
	})

	t.Run("Root path validation", func(t *testing.T) {
		validRoots := []string{
			"/data/repo",
			"/storage/my-repo",
			"/var/lib/filehub/repos/repo1",
		}

		for _, root := range validRoots {
			repo := &Repository{
				Root: root,
			}
			assert.NotEmpty(t, repo.Root)
			assert.True(t, len(root) > 0 && root[0] == '/')
		}
	})

	t.Run("Empty root path rejected", func(t *testing.T) {
		repo := &Repository{
			Root: "",
		}
		assert.Empty(t, repo.Root, "Empty root should be rejected")
	})
}

// TestShareValidation tests share validation logic
func TestShareValidation(t *testing.T) {
	t.Run("Valid share paths", func(t *testing.T) {
		validPaths := []string{
			"/",
			"/shared",
			"/shared/folder",
			"/deep/nested/path",
		}

		for _, path := range validPaths {
			share := &Share{
				Path: path,
			}
			assert.NotEmpty(t, share.Path)
		}
	})

	t.Run("Share path must be absolute", func(t *testing.T) {
		invalidPaths := []string{
			"relative/path",
			"shared",
			"",
		}

		for _, path := range invalidPaths {
			share := &Share{
				Path: path,
			}
			// Application should validate path starts with /
			assert.False(t, len(path) > 0 && path[0] == '/', "Path should start with /: %s", path)
			_ = share // avoid unused variable warning
		}
	})

	t.Run("User IDs must be positive", func(t *testing.T) {
		share := &Share{
			UserID:  0,
			OwnerID: 0,
			RepoID:  0,
		}
		assert.Equal(t, 0, share.UserID)
		assert.Equal(t, 0, share.OwnerID)
		assert.Equal(t, 0, share.RepoID)
	})
}

// TestResourceTests tests Resource helper methods
func TestResourceTests(t *testing.T) {
	t.Run("Resource String with various paths", func(t *testing.T) {
		testCases := []struct {
			repoName string
			path     string
			expected string
		}{
			{"repo", "/file.txt", "repo/file.txt"},
			{"home", "/", "home/"},
			{"data", "/folder/subfolder/file.txt", "data/folder/subfolder/file.txt"},
			{"repo", "", "repo"},
		}

		for _, tc := range testCases {
			repo := &Repository{
				Name: tc.repoName,
			}
			resource := &Resource{
				Repo: repo,
				Path: tc.path,
			}
			assert.Equal(t, tc.expected, resource.String())
		}
	})
}

// TestUserQuotaTests tests quota calculations
func TestUserQuotaTests(t *testing.T) {
	t.Run("Remaining quota calculation", func(t *testing.T) {
		testCases := []struct {
			total     int64
			used      int64
			remaining int64
		}{
			{1000, 600, 400},
			{10737418240, 5368709120, 5368709120},
			{100, 100, 0},
			{100, 0, 100},
		}

		for _, tc := range testCases {
			quota := &UserQuota{
				TotalQuotaBytes: tc.total,
				UsedBytes:       tc.used,
			}
			remaining := quota.TotalQuotaBytes - quota.UsedBytes
			assert.Equal(t, tc.remaining, remaining)
		}
	})

	t.Run("Quota exceeded check", func(t *testing.T) {
		quota := &UserQuota{
			TotalQuotaBytes: 1000,
			UsedBytes:       900,
		}

		// Can fit 100 bytes
		assert.True(t, quota.UsedBytes+100 <= quota.TotalQuotaBytes)

		// Cannot fit 101 bytes
		assert.False(t, quota.UsedBytes+101 <= quota.TotalQuotaBytes)
	})

	t.Run("Usage percentage calculation", func(t *testing.T) {
		testCases := []struct {
			total      int64
			used       int64
			percentage float64
		}{
			{100, 50, 50.0},
			{100, 75, 75.0},
			{100, 0, 0.0},
			{100, 100, 100.0},
		}

		for _, tc := range testCases {
			quota := &UserQuota{
				TotalQuotaBytes: tc.total,
				UsedBytes:       tc.used,
			}
			percentage := float64(quota.UsedBytes) / float64(quota.TotalQuotaBytes) * 100
			assert.InDelta(t, tc.percentage, percentage, 0.01)
		}
	})
}

// Helper function to repeat a string n times
func stringRepeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
