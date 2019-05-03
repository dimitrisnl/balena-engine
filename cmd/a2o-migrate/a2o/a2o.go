package a2o // import "github.com/balena-os/balena-engine/cmd/a2o-migrate/a2o"

import (
	"fmt"
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

	logrus.Infof("moving %d layer(s) to overlay", len(state.Layers))

	// move to overlay filetree
	for _, layer := range state.Layers {
		logrus := logrus.WithField("layer_id", layer.ID)

		var (
			layerDir = filepath.Join(overlayRoot, layer.ID)
		)

		logrus.Debugf("creating base dir %s", layerDir)
		// create /:layer_id dir
		err := os.MkdirAll(layerDir, 0700)
		if err != nil {
			return errors.Errorf("Error creating layer directory at %s: %w", layerDir, err)
		}

		logrus.Debug("creating layer link")
		// create /:layer_id/link file and /l/:layer_ref file
		_, err = overlayutil.CreateLayerLink(overlayRoot, layer.ID)
		if err != nil {
			return errors.Errorf("Error creating layer link dir: %w", err)
		}

		// create /:layer_id/lower
		for _, parentID := range layer.ParentIDs {
			logrus.Warn("todo: process parent layers")

		}

		layerDiffDir := filepath.Join(layerDir, "diff")
		aufsLayerDir := filepath.Join(aufsRoot, "diff", layer.ID)

		// migrate metadata files
		logrus.Debugf("processing metadata %d file(s)", len(layer.Meta))
		for _, meta := range layer.Meta {
			metaPath := filepath.Join(aufsLayerDir, meta.Path)

			logrus.WithField("meta_type", fmt.Sprintf("%v", meta.Type)).Debugf("translating %s to overlay", meta.Path)
			switch meta.Type {
			case MetaOpaque:
				// TODO set the opque xattr

			case MetaWhiteout:
				// TODO create the 0x0 char device

			case MetaOther:
				metaDir, metaFile := filepath.Split(metaPath)
				aufsMetaPath := filepath.Join(metaDir, aufsutil.WhiteoutMetaPrefix+metaFile)
				err = os.Remove(aufsMetaPath)
				if err != nil {
					return errors.Errorf("Error removing file at %s: %w", aufsMetaPath, err)
				}
			}
		}

		logrus.Info("moving aufs data to overlay")
		// move data over
		err = os.Rename(aufsLayerDir, layerDiffDir)
		if err != nil {
			return errors.Errorf("Error moving layer data from %s to %s: %w", aufsLayerDir, layerDiffDir, err)
		}
	}

	logrus.Warn("image migration not done yet!")
	// mv "$DOCKERDIR/image/aufs" "$DOCKERDIR/image/overlay2"

	logrus.Warn("container migration not done yet!")
	// # containers
	// log "Migrating containers ..."
	// for container_path in "$DOCKERDIR"/containers/*; do
	// 	container="$(basename "$container_path")"
	// 	log "---> $container"
	// 	jq '.Driver="overlay2"' "$container_path/config.v2.json" > "/tmp/a2o-migrate-container-$container.tmp"
	// 	mv "/tmp/a2o-migrate-container-$container.tmp" "$container_path/config.v2.json"
	// done

	logrus.Warn("daemon migration not done yet!")
	// sed -i "s/aufs/overlay2/g" /lib/systemd/system/balena.service
	// sed -i "s/aufs/overlay2/g" /etc/systemd/system/balena.service.d/balena.conf

	logrus.Debug("finished migration")
	return nil
}
