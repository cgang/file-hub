package stor

import (
	"io"
	"os"
)

// Storage defines an interface for file operations
type Storage interface {
	CreateFile(path string) (*os.File, error)
	DeleteFile(path string) error
	CreateDir(path string) error
	CopyFile(src, dst string) error
	MoveFile(src, dst string) error
	Exists(path string) (bool, error)
	GetFileInfo(path string) (os.FileInfo, error)
}

// OsStorage implements Storage using standard OS operations
type OsStorage struct{}

func (s *OsStorage) CreateFile(path string) (*os.File, error) {
	return os.Create(path)
}

func (s *OsStorage) DeleteFile(path string) error {
	return os.RemoveAll(path)
}

func (s *OsStorage) CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func (s *OsStorage) CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *OsStorage) MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}

func (s *OsStorage) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *OsStorage) GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
