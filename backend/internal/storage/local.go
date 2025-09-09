package storage

import (
	"io"
	"os"
	"path/filepath"
)

type Storage interface {
	Save(localTmpPath, destName string) (string, error)
}

type LocalStorage struct {
	basePath string
}

func NewLocal(basePath string) Storage {
	_ = os.MkdirAll(basePath, 0o755)
	return &LocalStorage{basePath: basePath}
}

func (l *LocalStorage) Save(tmpPath, destName string) (string, error) {
	dst := filepath.Join(l.basePath, destName)

	srcF, err := os.Open(tmpPath)
	if err != nil {
		return "", err
	}
	defer srcF.Close()

	dstF, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstF.Close()

	if _, err := io.Copy(dstF, srcF); err != nil {
		return "", err
	}

	return dst, nil
}
