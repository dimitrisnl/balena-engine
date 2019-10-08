package a2o

import (
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// FailCleanup should be run after a failed migration.
// It will remove any files left over from the migration process
// and migrate containers back to aufs.
//
func FailCleanup() error {
	logrus.Info("recovering from failed migration")

	err := removeDirIfExists(tempTargetRoot())
	if err != nil {
		return err
	}

	err = removeDirIfExists(overlayRoot())
	if err != nil {
		return err
	}

	overlayImageDir := filepath.Join(StorageRoot, "image", "overlay2")
	err = removeDirIfExists(overlayImageDir)
	if err != nil {
		return err
	}

	err = SwitchAllContainersStorageDriver("aufs")
	if err != nil {
		return fmt.Errorf("Error migrating containers to aufs: %v", err)
	}

	return nil
}
