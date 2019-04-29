package a2o // import "github.com/balena-os/balena-engine/cmd/a2o-migrate/a2o"

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	errors "golang.org/x/xerrors"
)

const (
	dockerDir = "/var/run/balena-engine"
)

var (
	aufsRoot     = filepath.Join(dockerDir, "aufs")
	overlay2Root = filepath.Join(dockerDir, "overlay2")
)

func AuFSToOverlay2() error {
	logrus.Debug("starting a2o migration")

	var err error

	// make sure we actually have an aufs tree to migrate from
	err = checkAufsExists(aufsRoot)
	if err != nil {
		return err
	}

	// make sure there isn't an overlay2 tree already
	err = checkOverlay2NotExists(overlay2Root)
	if err != nil {
		return err
	}

	err = filepath.Walk(aufsRoot, filepath.WalkFunc(processor))
	if err != nil {
		return errors.Errorf("failed to walk aufs tree: %w", err)
	}

	logrus.Debug("finished a2o migration")
	return nil
}

func processor(path string, fi os.FileInfo, err error) error {
	logrus.Debug(path)
	return nil
}
