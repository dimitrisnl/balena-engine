package a2o

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/balena-os/balena-engine/cmd/a2o-migrate/osutil"
)

// Rollback should be run after a unsuccesful migration.
// It will remove any files left over from the migration process and
// reconfigure balenaEngine to use aufs again.
//
func Rollback() error {
	logrus.Info("starting overlay2 -> aufs rollback")
	logrus.Warnf("rolling back to aufs, this removes %s if it exists", overlayRoot)

	err := removeIfExists(tempTargetRoot, true)
	if err != nil {
		return err
	}

	err = removeIfExists(overlayRoot, true)
	if err != nil {
		return err
	}

	return nil
}

func removeIfExists(path string, isDir bool) error {
	ok, err := osutil.Exists(path, isDir)
	if err != nil {
		return err
	}
	if ok {
		logrus.Infof("removing %s", path)
		err = os.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}
