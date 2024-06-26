package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func GetRootPath() string {
	result, err := filepath.Abs(viper.GetString("storagePath"))
	if err != nil {
		logrus.Fatalln("Failed to get storage path:", err)
	}
	return result
}

func GetTmpPath() string {
	return filepath.Join(GetRootPath(), "tmp")
}

func GetCachePath() string {
	return filepath.Join(GetRootPath(), "cache")
}

func Initialize() {
	err := os.MkdirAll(GetTmpPath(), 0700)
	if err != nil {
		logrus.Fatalln("Failed to create tmp dir:", err)
	}
	err = os.MkdirAll(GetCachePath(), 0700)
	if err != nil {
		logrus.Fatalln("Failed to create cache dir:", err)
	}
}

func CreateTemp(pattern string) (*os.File, error) {
	return os.CreateTemp(GetTmpPath(), pattern)
}

func MkdirTemp(pattern string) (string, error) {
	return os.MkdirTemp(GetTmpPath(), pattern)
}

func DownloadFile(ctx context.Context, url string, hash string) error {
	// get tmp file
	file, err := CreateTemp("download-" + hash + "-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()

	// Download url to tmp file
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	n, err := io.Copy(file, res.Body)
	if err != nil {
		return err
	}
	logrus.Println("Downloaded", n, "bytes")

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	// Calculate sha256 hash
	hasher := sha256.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return err
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))
	if fileHash != hash {
		logrus.Println("Hash mismatch:", fileHash, "!=", hash)
		return fmt.Errorf("file hash mismatch")
	}

	// Move tmp file to cache
	cachePath := GetCachePath()
	filePath := filepath.Join(cachePath, hash)
	return os.Rename(file.Name(), filePath)
}

func PrepareFile(ctx context.Context, url string, hash string) (string, error) {
	logrus.Println("Preparing file:", url, hash)
	cachePath := GetCachePath()
	filePath := filepath.Join(cachePath, hash)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Download file with retry
		retry := 3
		for retry > 0 {
			err := DownloadFile(ctx, url, hash)
			if err != nil {
				logrus.Println("Failed to download file:", err)
				retry--
				continue
			}
			break
		}
	}
	return filePath, nil
}
