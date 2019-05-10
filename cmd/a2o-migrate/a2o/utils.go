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
	containerConfigPath := filepath.Join(balenaEngineDir, "containers", containerID, "config.v2.json")
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
