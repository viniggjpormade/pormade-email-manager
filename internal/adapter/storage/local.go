package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/viniggjpormade/pormade-email-manager/internal/domain"
)

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) domain.StorageProvider {
	os.MkdirAll(baseDir, os.ModePerm)
	return &LocalStorage{baseDir: baseDir}
}

func (l *LocalStorage) Upload(filename string, data []byte) (string, error) {
	uniqueName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)
	filePath := filepath.Join(l.baseDir, uniqueName)
	
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return "", err
	}
	
	return filePath, nil
}
