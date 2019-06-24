package a2o

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Rollback should be run after a unsuccesful migration.
// It will remove any files left over from the migration process and
// reconfigure balenaEngine to use aufs again.
//
func Rollback() error {
	logrus.Info("starting overlay2 -> aufs rollback")

	err := removeDirIfExists(tempTargetRoot)
	if err != nil {
		return err
	}

	err = removeDirIfExists(overlayRoot)
	if err != nil {
		return err
	}

	overlayImageDir := filepath.Join(StorageRoot, "image", "overlay2")
	err = removeDirIfExists(overlayImageDir)
	if err != nil {
		return err
	}

	return nil
}
