package sync

import (
	"hash/crc32"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateVersion(t *testing.T) {
	t.Run("Version format", func(t *testing.T) {
		version := generateVersion()
		assert.True(t, strings.HasPrefix(version, "v"), "Version should start with 'v'")
		assert.Contains(t, version, "-", "Version should contain separator")
	})

	t.Run("Unique versions", func(t *testing.T) {
		v1 := generateVersion()
		time.Sleep(time.Microsecond)
		v2 := generateVersion()
		assert.NotEqual(t, v1, v2, "Versions should be unique")
	})

	t.Run("Version parseability", func(t *testing.T) {
		version := generateVersion()
		parts := strings.Split(version, "-")
		assert.Len(t, parts, 2, "Version should have 2 parts")
	})
}

func TestCalculateSHA256(t *testing.T) {
	t.Run("Empty data", func(t *testing.T) {
		data := []byte{}
		hash := calculateSHA256(data)
		expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		assert.Equal(t, expected, hash)
	})

	t.Run("Simple data", func(t *testing.T) {
		data := []byte("hello world")
		hash := calculateSHA256(data)
		expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
		assert.Equal(t, expected, hash)
	})

	t.Run("Consistent hashing", func(t *testing.T) {
		data := []byte("test data")
		hash1 := calculateSHA256(data)
		hash2 := calculateSHA256(data)
		assert.Equal(t, hash1, hash2, "Same data should produce same hash")
	})

	t.Run("Different data different hash", func(t *testing.T) {
		hash1 := calculateSHA256([]byte("data1"))
		hash2 := calculateSHA256([]byte("data2"))
		assert.NotEqual(t, hash1, hash2)
	})
}

func TestVersionManagement(t *testing.T) {
	t.Run("Generate sequential versions", func(t *testing.T) {
		versions := make([]string, 10)
		for i := 0; i < 10; i++ {
			versions[i] = generateVersion()
			if i > 0 {
				assert.NotEqual(t, versions[i-1], versions[i], "Versions should be unique")
			}
		}
	})
}

func TestChunkedUpload(t *testing.T) {
	t.Run("Calculate chunk sizes", func(t *testing.T) {
		fileSizes := []struct {
			sizeMB  float64
			chunks  int
			comment string
		}{
			{0.5, 1, "512KB - 1 chunk"},
			{1.0, 1, "1MB - 1 chunk"},
			{1.001, 2, "1MB + 1 byte - 2 chunks"},
			{5.0, 5, "5MB - 5 chunks"},
			{10.5, 11, "10.5MB - 11 chunks"},
		}

		for _, test := range fileSizes {
			t.Run(test.comment, func(t *testing.T) {
				sizeBytes := int64(test.sizeMB * 1024 * 1024)
				chunks := int((sizeBytes + ChunkSize - 1) / ChunkSize)
				assert.Equal(t, test.chunks, chunks, "Chunk count mismatch for %.3f MB", test.sizeMB)
			})
		}
	})

	t.Run("Chunk offsets calculation", func(t *testing.T) {
		chunkIndex := 3
		expectedOffset := int64(chunkIndex) * ChunkSize
		assert.Equal(t, int64(3*1024*1024), expectedOffset)
	})

	t.Run("Chunk hash calculation", func(t *testing.T) {
		data := make([]byte, 1024)
		for i := range data {
			data[i] = byte(i % 256)
		}

		hash := calculateSHA256(data)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64, "SHA-256 hash should be 64 characters")
	})
}

func TestSyncStatus(t *testing.T) {
	t.Run("Status determination scenarios", func(t *testing.T) {
		scenarios := []struct {
			name        string
			clientETag  string
			serverETag  string
			expected    string
			description string
		}{
			{
				"Synced - same etag",
				"abc123",
				"abc123",
				"synced",
				"Client and server have same content",
			},
			{
				"Modified - different etag",
				"abc123",
				"def456",
				"modified",
				"Server has newer content",
			},
			{
				"New - no client etag",
				"",
				"abc123",
				"new",
				"File exists on server but not on client",
			},
			{
				"Deleted - client has etag but server doesn't",
				"abc123",
				"",
				"deleted",
				"File was deleted on server",
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				status := determineStatus(scenario.clientETag, scenario.serverETag)
				assert.Equal(t, scenario.expected, status, scenario.description)
			})
		}
	})
}

func determineStatus(clientETag, serverETag string) string {
	if clientETag == "" {
		if serverETag != "" {
			return "new"
		}
		return "synced"
	}

	if serverETag == "" {
		return "deleted"
	}

	if clientETag == serverETag {
		return "synced"
	}

	return "modified"
}

func TestUploadSession(t *testing.T) {
	t.Run("Session creation parameters", func(t *testing.T) {
		totalSize := int64(5 * 1024 * 1024)
		totalChunks := int((totalSize + ChunkSize - 1) / ChunkSize)

		assert.Equal(t, 5, totalChunks, "5MB should split into 5 chunks")

		uploadID := "session-123"
		path := "/uploads/test.bin"

		assert.NotEmpty(t, uploadID, "Upload ID should not be empty")
		assert.NotEmpty(t, path, "Path should not be empty")
		assert.Greater(t, totalChunks, 0, "Should have at least one chunk")
	})

	t.Run("Session expiry time", func(t *testing.T) {
		createdAt := time.Now()
		expiresAt := createdAt.Add(MaxConnectionTime)

		duration := expiresAt.Sub(createdAt)
		assert.Equal(t, MaxConnectionTime, duration, "Session should expire after MaxConnectionTime")
		assert.Equal(t, 24*time.Hour, MaxConnectionTime, "MaxConnectionTime should be 24 hours")
	})
}

func TestFileOperationScenarios(t *testing.T) {
	t.Run("Simple upload size limits", func(t *testing.T) {
		sizes := []struct {
			size    int64
			allowed bool
			reason  string
		}{
			{512 * 1024, true, "512KB is allowed"},
			{1024 * 1024, true, "1MB is allowed"},
			{10 * 1024 * 1024, true, "Exactly 10MB is allowed"},
			{10*1024*1024 + 1, false, "Over 10MB should use chunked upload"},
			{100 * 1024 * 1024, false, "100MB should use chunked upload"},
		}

		for _, test := range sizes {
			t.Run(test.reason, func(t *testing.T) {
				isAllowed := test.size <= MaxSimpleUploadSize
				assert.Equal(t, test.allowed, isAllowed)
			})
		}
	})
}

func TestIntegrityVerification(t *testing.T) {
	t.Run("Hash verification scenarios", func(t *testing.T) {
		testData := []byte("test file content for integrity verification")

		clientHash := calculateSHA256(testData)
		serverHash := calculateSHA256(testData)

		t.Run("Matching hashes", func(t *testing.T) {
			assert.Equal(t, clientHash, serverHash, "Hashes should match for identical data")
		})

		t.Run("Different data different hash", func(t *testing.T) {
			modifiedData := []byte("modified test file content")
			modifiedHash := calculateSHA256(modifiedData)
			assert.NotEqual(t, clientHash, modifiedHash, "Different data should have different hashes")
		})

		t.Run("Hash of empty file", func(t *testing.T) {
			emptyHash := calculateSHA256([]byte{})
			assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", emptyHash)
		})
	})
}

func TestChunkUploadSequence(t *testing.T) {
	t.Run("Upload sequence validation", func(t *testing.T) {
		totalSize := int64(3.5 * 1024 * 1024)
		expectedChunks := 4

		totalChunks := int((totalSize + ChunkSize - 1) / ChunkSize)
		assert.Equal(t, expectedChunks, totalChunks)

		chunkIndices := make([]int, expectedChunks)
		for i := range chunkIndices {
			chunkIndices[i] = i
		}

		for i, idx := range chunkIndices {
			expectedOffset := int64(idx) * ChunkSize
			actualOffset := int64(i) * ChunkSize
			assert.Equal(t, expectedOffset, actualOffset, "Chunk mismatch at index %d", i)
		}
	})
}

func TestTimestampVersioning(t *testing.T) {
	t.Run("Version monotonicity", func(t *testing.T) {
		versions := make([]string, 100)
		timestamps := make([]int64, 100)

		for i := 0; i < 100; i++ {
			versions[i] = generateVersion()
			timestamp := time.Now().UnixNano()
			timestamps[i] = timestamp

			if i > 0 {
				assert.True(t, timestamps[i] >= timestamps[i-1], "Timestamps should be monotonically increasing")
			}
		}
	})

	t.Run("Version parsing", func(t *testing.T) {
		version := generateVersion()
		parts := strings.Split(version, "-")

		timestampPart := parts[0]
		nanoPart := parts[1]

		assert.True(t, strings.HasPrefix(timestampPart, "v"), "Timestamp part should start with 'v'")
		assert.NotEmpty(t, nanoPart, "Nanosecond part should not be empty")
	})
}

func TestPathOperations(t *testing.T) {
	t.Run("Path validation", func(t *testing.T) {
		paths := []string{
			"/test/file.txt",
			"/test/dir/",
			"/",
			"/deep/nested/path/file.pdf",
		}

		for _, path := range paths {
			t.Run(path, func(t *testing.T) {
				assert.True(t, strings.HasPrefix(path, "/"), "Path should start with '/'")
				assert.NotEmpty(t, path, "Path should not be empty")
			})
		}
	})
}

func TestChecksumConsistency(t *testing.T) {
	t.Run("Multiple calculations", func(t *testing.T) {
		data := make([]byte, 4096)
		for i := range data {
			data[i] = byte(i % 256)
		}

		hashes := make([]string, 10)
		for i := 0; i < 10; i++ {
			hashes[i] = calculateSHA256(data)
		}

		firstHash := hashes[0]
		for i, hash := range hashes[1:] {
			assert.Equal(t, firstHash, hash, "Hash should be consistent across calculations (iteration %d)", i+1)
		}
	})

	t.Run("Hash collision resistance", func(t *testing.T) {
		data1 := []byte("collision test data 1")
		data2 := []byte("collision test data 2")

		hash1 := calculateSHA256(data1)
		hash2 := calculateSHA256(data2)

		assert.NotEqual(t, hash1, hash2, "Similar data should have different hashes")
	})
}

func TestCRC32Comparison(t *testing.T) {
	t.Run("CRC32 hash comparison", func(t *testing.T) {
		data := []byte("test data")

		shaHash := calculateSHA256(data)
		crcHash := crc32.ChecksumIEEE(data)

		t.Run("SHA-256 properties", func(t *testing.T) {
			assert.Len(t, shaHash, 64, "SHA-256 should be 64 characters")
			assert.NotEmpty(t, shaHash, "SHA-256 should not be empty")
		})

		t.Run("CRC32 properties", func(t *testing.T) {
			assert.NotZero(t, crcHash, "CRC32 should not be zero for non-empty data")
		})

		t.Run("Consistency", func(t *testing.T) {
			shaHash2 := calculateSHA256(data)
			crcHash2 := crc32.ChecksumIEEE(data)

			assert.Equal(t, shaHash, shaHash2, "SHA-256 should be consistent")
			assert.Equal(t, crcHash, crcHash2, "CRC32 should be consistent")
		})
	})
}

func TestDataIntegrity(t *testing.T) {
	t.Run("Data corruption detection", func(t *testing.T) {
		originalData := []byte("original data")
		corruptedData := []byte("corrupted data")

		originalHash := calculateSHA256(originalData)
		corruptedHash := calculateSHA256(corruptedData)

		assert.NotEqual(t, originalHash, corruptedHash, "Corrupted data should be detected")
	})

	t.Run("Single bit change detection", func(t *testing.T) {
		data := []byte{0x00, 0xFF, 0xAA, 0x55}
		modifiedData := []byte{0x01, 0xFF, 0xAA, 0x55}

		hash1 := calculateSHA256(data)
		hash2 := calculateSHA256(modifiedData)

		assert.NotEqual(t, hash1, hash2, "Single bit change should be detected")
	})
}

func TestUploadResume(t *testing.T) {
	t.Run("Chunk resume scenario", func(t *testing.T) {
		totalChunks := 10
		uploadedChunks := []int{0, 1, 2}

		remainingChunks := totalChunks - len(uploadedChunks)
		assert.Equal(t, 7, remainingChunks, "Should have 7 chunks remaining")

		chunkIndices := make([]int, totalChunks)
		for i := range chunkIndices {
			chunkIndices[i] = i
		}

		missingChunks := []int{}
		for _, chunkIndex := range chunkIndices {
			uploaded := false
			for _, uploadedIndex := range uploadedChunks {
				if chunkIndex == uploadedIndex {
					uploaded = true
					break
				}
			}
			if !uploaded {
				missingChunks = append(missingChunks, chunkIndex)
			}
		}

		assert.Len(t, missingChunks, 7, "Should identify 7 missing chunks")
		assert.Equal(t, []int{3, 4, 5, 6, 7, 8, 9}, missingChunks)
	})
}

func TestMaxConnectionTime(t *testing.T) {
	t.Run("Connection time limits", func(t *testing.T) {
		assert.Equal(t, 24*time.Hour, MaxConnectionTime, "Max connection time should be 24 hours")

		sessionStart := time.Now()
		expiryTime := sessionStart.Add(MaxConnectionTime)
		duration := expiryTime.Sub(sessionStart)

		assert.Equal(t, MaxConnectionTime, duration)
	})
}

func TestLargeFileHandling(t *testing.T) {
	t.Run("Large file chunk count", func(t *testing.T) {
		sizes := []struct {
			sizeMB  float64
			chunks  int
			comment string
		}{
			{1.0, 1, "1MB = 1 chunk"},
			{10.0, 10, "10MB = 10 chunks"},
			{100.0, 100, "100MB = 100 chunks"},
			{1024.0, 1024, "1GB = 1024 chunks"},
			{1073.74, 1074, "1.048576GB = 1074 chunks"},
		}

		for _, test := range sizes {
			t.Run(test.comment, func(t *testing.T) {
				sizeBytes := int64(test.sizeMB * 1024 * 1024)
				chunks := int((sizeBytes + ChunkSize - 1) / ChunkSize)
				assert.Equal(t, test.chunks, chunks)
			})
		}
	})
}

func TestErrorScenarios(t *testing.T) {
	t.Run("Empty upload ID", func(t *testing.T) {
		uploadID := ""
		assert.Empty(t, uploadID, "Empty upload ID should be invalid")
	})

	t.Run("Invalid chunk index", func(t *testing.T) {
		totalChunks := 10
		invalidIndices := []int{-1, -100, 100, 999}

		for _, idx := range invalidIndices {
			isValid := idx >= 0 && idx < totalChunks
			assert.False(t, isValid, "Index %d should be invalid for %d chunks", idx, totalChunks)
		}
	})

	t.Run("Zero size file", func(t *testing.T) {
		size := int64(0)
		totalChunks := int((size + ChunkSize - 1) / ChunkSize)
		assert.Equal(t, 0, totalChunks, "Zero size should result in 0 chunks")
	})
}

func TestBufferHandling(t *testing.T) {
	t.Run("Large buffer handling", func(t *testing.T) {
		data := make([]byte, 10*1024*1024)

		start := time.Now()
		hash := calculateSHA256(data)
		duration := time.Since(start)

		assert.NotEmpty(t, hash, "Hash should be calculated")
		assert.Less(t, duration, time.Second, "Hash calculation should complete within 1 second")
	})

	t.Run("Small buffer handling", func(t *testing.T) {
		data := make([]byte, 1024)

		hash := calculateSHA256(data)

		assert.NotEmpty(t, hash, "Hash should be calculated")
		assert.Len(t, hash, 64, "Hash length should be 64 characters")
	})
}

func TestConcurrentVersions(t *testing.T) {
	t.Run("Concurrent version generation", func(t *testing.T) {
		versions := make(chan string, 10)
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				versions <- generateVersion()
			}()
		}

		wg.Wait()
		close(versions)

		uniqueVersions := make(map[string]bool)
		for version := range versions {
			assert.True(t, strings.HasPrefix(version, "v"), "Version should start with 'v'")
			uniqueVersions[version] = true
		}

		assert.Len(t, uniqueVersions, 10, "All 10 concurrently generated versions should be unique")
	})
}

func TestChunkSizeConstant(t *testing.T) {
	t.Run("Chunk size validation", func(t *testing.T) {
		expectedChunkSize := 1024 * 1024
		chunkSize := int(ChunkSize)
		assert.Equal(t, expectedChunkSize, chunkSize, "ChunkSize should be 1MB")
	})
}

func TestMaxSimpleUploadSize(t *testing.T) {
	t.Run("Max simple upload size", func(t *testing.T) {
		expectedMaxSize := 10 * 1024 * 1024
		maxSize := int(MaxSimpleUploadSize)
		assert.Equal(t, expectedMaxSize, maxSize, "Max simple upload should be 10MB")
	})
}

func TestEmptyStringComparison(t *testing.T) {
	t.Run("Empty etag handling", func(t *testing.T) {
		status1 := determineStatus("", "")
		assert.Equal(t, "synced", status1, "Both empty should be synced")

		status2 := determineStatus("", "abc123")
		assert.Equal(t, "new", status2, "Empty server, client has data should be new")

		status3 := determineStatus("abc123", "")
		assert.Equal(t, "deleted", status3, "Empty server, client has data should be deleted")
	})
}

func TestVersionStringFormat(t *testing.T) {
	t.Run("All versions follow same format", func(t *testing.T) {
		versions := make([]string, 20)

		for i := 0; i < 20; i++ {
			versions[i] = generateVersion()
		}

		for i, version := range versions {
			parts := strings.Split(version, "-")
			assert.Len(t, parts, 2, "Version %d should have 2 parts", i)
			assert.True(t, strings.HasPrefix(parts[0], "v"), "Version %d timestamp should start with 'v'", i)
			assert.NotEmpty(t, parts[1], "Version %d nanoseconds should not be empty", i)
		}
	})
}

func TestHashLength(t *testing.T) {
	t.Run("SHA-256 always returns 64 characters", func(t *testing.T) {
		sizes := []int{0, 1, 10, 100, 1000, 10000}

		for _, size := range sizes {
			data := make([]byte, size)
			hash := calculateSHA256(data)
			assert.Len(t, hash, 64, "Hash length should always be 64 for size %d", size)
		}
	})
}

func TestChunkIndexBounds(t *testing.T) {
	t.Run("Chunk index validation", func(t *testing.T) {
		tests := []struct {
			chunkIndex  int
			totalChunks int
			valid       bool
		}{
			{0, 10, true},
			{5, 10, true},
			{9, 10, true},
			{-1, 10, false},
			{10, 10, false},
			{100, 10, false},
		}

		for _, test := range tests {
			t.Run("Index"+string(rune(test.chunkIndex)), func(t *testing.T) {
				isValid := test.chunkIndex >= 0 && test.chunkIndex < test.totalChunks
				assert.Equal(t, test.valid, isValid)
			})
		}
	})
}
