package a2o // import "github.com/balena-os/balena-engine/cmd/a2o-migrate/a2o"

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	errors "golang.org/x/xerrors"

	"github.com/balena-os/balena-engine/cmd/a2o-migrate/aufsutil"
	"github.com/balena-os/balena-engine/cmd/a2o-migrate/overlayutil"
)

const (
	balenaEngineDir = "/var/lib/balena-engine"
)

var (
	aufsRoot    = filepath.Join(balenaEngineDir, "aufs")
	overlayRoot = filepath.Join(balenaEngineDir, "overlay2")
)

// AuFSToOverlay migrates the state of the storage from aufs -> overlay2
func AuFSToOverlay() error {
	logrus.Debug("starting a2o migration")

	var err error

	// make sure we actually have an aufs tree to migrate from
	err = aufsutil.CheckRootExists(balenaEngineDir)
	if err != nil {
		return err
	}

	// make sure there isn't an overlay2 tree already
	err = overlayutil.CheckRootExists(balenaEngineDir)
	if err == nil {
		return errors.New("Overlay2 directory exists, not overwriting")
	}

	var state State

	diffDir := filepath.Join(aufsRoot, "diff")

	// get all layers
	layerIDs, err := aufsutil.LoadFiles(diffDir)
	if err != nil {
		return errors.Errorf("Error loading layer ids: %w", err)
	}
	logrus.Debugf("layer ids in %s: %+#v", diffDir, layerIDs)

	for _, layerID := range layerIDs {
		logrus := logrus.WithField("layer_id", layerID)
		logrus.Debug("parsing layer")
		layer := Layer{ID: layerID}

		// get parent layers
		logrus.Debug("parsing parent ids")
		parentIDs, err := aufsutil.GetParentIDs(aufsRoot, layerID)
		if err != nil {
			return errors.Errorf("Error loading parent IDs for %s: %w", layerID, err)
		}
		layer.ParentIDs = parentIDs

		layerDir := filepath.Join(diffDir, layerID)
		logrus.Debug("parsing for metadata files")
		err = filepath.Walk(layerDir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			r := strings.SplitAfter(path, layerDir)
			if len(r) != 2 {
				return errors.Errorf("unexpected path: %s", path)
			}
			absPath := r[1]
			logrus := logrus.WithField("path", absPath)

			if !fi.IsDir() && aufsutil.IsWhiteout(fi.Name()) {
				if aufsutil.IsWhiteoutMeta(fi.Name()) {
					if aufsutil.IsOpaqueParentDir(fi.Name()) {
						logrus.Debug("discovered opaque-dir marker")
						layer.Meta = append(layer.Meta, Meta{
							Path: filepath.Dir(absPath),
							Type: MetaOpaque,
						})
						return nil
					}

					logrus.Debug("discovered whiteout-meta marker")
					// other whiteout metadata
					layer.Meta = append(layer.Meta, Meta{
						Path: absPath,
						Type: MetaOther,
					})
				}

				logrus.Debug("discovered whiteout marker")
				// simple whiteout file
				layer.Meta = append(layer.Meta, Meta{
					Path: filepath.Join(filepath.Dir(absPath), aufsutil.StripWhiteoutPrefix(fi.Name())),
					Type: MetaWhiteout,
				})
			}
			return nil
		})
		if err != nil {
			return errors.Errorf("Error walking filetree in %s: %w", layerDir, err)
		}

		state.Layers = append(state.Layers, layer)
		logrus.Debug("done")
	}

	logrus.Debugf("final state %#+v", state)
	logrus.Debug("finished a2o migration")
	return nil
}
