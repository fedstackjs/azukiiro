package utils

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fedstackjs/azukiiro/storage"
	"github.com/sirupsen/logrus"
)

func UnzipTemp(source string, target string) (string, error) {
	dir, err := storage.MkdirTemp(target)
	if err != nil {
		return dir, err
	}
	logrus.Infof("Unzipping %s to %s", source, dir)
	err = exec.Command("unzip", source, "-d", dir).Run()
	if err != nil {
		os.RemoveAll(dir)
		logrus.Println("Error unzipping", source, ":", err)
		return dir, fmt.Errorf("failed to extract solution")
	}
	// remove all symlinks to avoid security issues
	err = exec.Command("find", dir, "-type", "l", "-delete").Run()
	if err != nil {
		os.RemoveAll(dir)
		logrus.Println("Error removing symlinks in", dir, ":", err)
		return dir, fmt.Errorf("failed to extract solution")
	}
	return dir, nil
}
