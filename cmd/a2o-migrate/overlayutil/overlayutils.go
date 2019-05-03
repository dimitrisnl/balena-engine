package overlayutil

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
	errors "golang.org/x/xerrors"

	"github.com/balena-os/balena-engine/cmd/a2o-migrate/osutil"
)

var (
	// ErrOverlayRootNotExists indicates the overlay2 root directory wasn't found
	ErrOverlayRootNotExists = errors.New("Overlay2 root doesn't exists")
)

// CheckRootExists checks for the overlay storage root directory
func CheckRootExists(engineDir string) error {
	root := filepath.Join(engineDir, "overlay2")
	logrus.WithField("overlay_root", root).Debug("checking if overlay2 root exists")
	ok, err := osutil.Exists(root, true)
	if err != nil {
		return err
	}
	if !ok {
		return ErrOverlayRootNotExists
	}
	return nil
}
