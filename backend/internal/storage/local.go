package storage

import (
	"io"
	"os"
	"path/filepath"
)

type Local struct{ base string }

func NewLocal(base string) *Local {
	_ = os.MkdirAll(base, 0o755)
	return &Local{base: base}
}

func (l *Local) Save(tmpPath, destName string) (string, error) {
	dst := filepath.Join(l.base, destName)

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
