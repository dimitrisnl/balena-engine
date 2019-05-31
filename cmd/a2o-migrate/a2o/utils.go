package a2o

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// switchContainerStorageDriver rewrites the container config to use a new storage driver,
// this is the only change needed to make it work after the migration
func switchContainerStorageDriver(containerID, newStorageDriver string) error {
	containerConfigPath := filepath.Join(StorageRoot, "containers", containerID, "config.v2.json")
	f, err := os.OpenFile(containerConfigPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	var containerConfig = make(map[string]interface{})
	err = json.NewDecoder(f).Decode(&containerConfig)
	if err != nil {
		return err
	}
	containerConfig["Driver"] = "overlay2"

	err = f.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(&containerConfig)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return err
	}
	return nil
}

// replicate hardlinks all files from sourceDir to targetDir, reusing the same
// file structure
func replicate(sourceDir, targetDir string) error {
	return filepath.Walk(sourceDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var (
			targetPath = strings.Replace(path, sourceDir, targetDir, 1)
			logrus     = logrus.WithField("path", targetPath)
		)

		if fi.IsDir() {
			logrus.Debug("creating directory")
			err = os.MkdirAll(targetPath, os.ModeDir|0755)
			if err != nil {
				return err
			}
		} else {
			logrus.Debug("create hardlink")
			err = os.Link(path, targetPath)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
