package a2o // import "github.com/balena-os/balena-engine/cmd/a2o-migrate/a2o"

import (
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func checkAufsExists(root string) error {
	logrus.WithField("aufs_root", root).Debug("checking if aufs root exists")
	fi, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrAuFSNotExists
		}
		return err
	}
	if !fi.IsDir() {
		return ErrAuFSNotExists
	}
	return nil
}

func checkOverlay2NotExists(root string) error {
	logrus.WithField("overlay2_root", root).Debug("checking if overlay2 not exists")
	fi, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if fi.IsDir() {
		return ErrOverlay2Exists
	}
	return errors.Errorf("%s exists and is not a directory", root)
}
