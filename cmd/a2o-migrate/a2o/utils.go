package a2o // import "github.com/balena-os/balena-engine/cmd/a2o-migrate/a2o"

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	errors "golang.org/x/xerrors"
)

var (
	ErrAuFSNotExists    = errors.New("aufs tree doesn't exists")
	ErrOverlayNotExists = errors.New("overlay2 tree doesn't exists")
)

func checkAufsExists(engineDir string) error {
	root := filepath.Join(engineDir, "aufs")
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

func checkOverlayExists(engineDir string) error {
	root := filepath.Join(engineDir, "overlay2")
	logrus.WithField("overlay_root", root).Debug("checking if overlay2 root exists")
	fi, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrOverlayNotExists
		}
		return err
	}
	if !fi.IsDir() {
		return ErrOverlayNotExists
	}
	return nil
}
