package a2o // import "github.com/balena-os/balena-engine/cmd/a2o-migrate/a2o"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	errors "golang.org/x/xerrors"

	"github.com/balena-os/balena-engine/cmd/a2o-migrate/aufsutil"
	"github.com/balena-os/balena-engine/cmd/a2o-migrate/osutil"
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
		logrus.Warn("overlay root found, cleaning up...")
		err := os.Remove(overlayRoot)
		if err != nil {
			return errors.Errorf("Error cleaning up %s: %w", overlayRoot, err)
		}
	}

	var (
		state State
	)

	diffDir := filepath.Join(aufsRoot, "diff")

	// get all layers
	layerIDs, err := osutil.LoadIDs(diffDir)
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
						layer.Meta = append(layer.Meta, Meta{
							Path: filepath.Dir(absPath),
							Type: MetaOpaque,
						})
						logrus.Debug("discovered opaque-dir marker")
						return nil
					}

					// other whiteout metadata
					layer.Meta = append(layer.Meta, Meta{
						Path: absPath,
						Type: MetaOther,
					})
					logrus.Debug("discovered whiteout-meta marker")
					return nil
				}

				// simple whiteout file
				layer.Meta = append(layer.Meta, Meta{
					Path: filepath.Join(filepath.Dir(absPath), aufsutil.StripWhiteoutPrefix(fi.Name())),
					Type: MetaWhiteout,
				})
				logrus.Debug("discovered whiteout marker")
				return nil
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

	var (
		tempTargetRoot = filepath.Join(balenaEngineDir, "overlay2.temp")
	)

	// move to overlay filetree
	for _, layer := range state.Layers {
		logrus := logrus.WithField("layer_id", layer.ID)

		var (
			layerDir = filepath.Join(tempTargetRoot, layer.ID)
		)

		logrus.Debugf("creating base dir %s", layerDir)
		// create /:layer_id dir
		err := os.MkdirAll(layerDir, 0700)
		if err != nil {
			return errors.Errorf("Error creating layer directory at %s: %w", layerDir, err)
		}

		logrus.Debug("creating layer link")
		// create /:layer_id/link file and /l/:layer_ref file
		_, err = overlayutil.CreateLayerLink(tempTargetRoot, layer.ID)
		if err != nil {
			return errors.Errorf("Error creating layer link dir: %w", err)
		}

		logrus.Debug("processing parent layers")
		// create /:layer_id/lower
		var lower string
		for _, parentID := range layer.ParentIDs {
			logrus := logrus.WithField("parent_layer_id", parentID)

			parentLayerDir := filepath.Join(tempTargetRoot, parentID)
			ok, err := osutil.Exists(parentLayerDir, true)
			if err != nil {
				return errors.Errorf("Error checking for %s: %w", parentLayerDir, err)
			}
			if !ok {
				// parent layer hasn't been processed separately yet.
				logrus.Debugf("creating parent layer base dir %s", parentLayerDir)
				err := os.MkdirAll(parentLayerDir, 0700)
				if err != nil {
					return errors.Errorf("Error creating layer directory at %s: %w", parentLayerDir, err)
				}
			}
			logrus.Debug("creating parent layer link")
			parentRef, err := overlayutil.CreateLayerLink(tempTargetRoot, parentID)
			if err != nil {
				return errors.Errorf("Error creating layer link dir for parent layer: %w", err)
			}
			lower = overlayutil.AppendLower(lower, parentRef)
		}
		if lower != "" {
			lowerFile := filepath.Join(layerDir, "lower")
			logrus.Debugf("creating lower at %s", lowerFile)
			err := ioutil.WriteFile(lowerFile, []byte(lower), 0644)
			if err != nil {
				return errors.Errorf("Error creating file at %s: %w", lowerFile, err)
			}
			layerWorkDir := filepath.Join(layerDir, "work")
			logrus.Debugf("creating work dir at %s", lowerFile)
			err = os.MkdirAll(layerWorkDir, 0700)
			if err != nil {
				return errors.Errorf("Error creating directory at %s: %w", layerWorkDir, err)
			}
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
				// set the opque xattr
				err := overlayutil.SetOpaque(metaPath)
				if err != nil {
					return errors.Errorf("Error marking %s as opque: %w", metaPath, err)
				}
				// remove aufs metadata file
				aufsMetaPath := filepath.Join(metaPath, aufsutil.OpaqueDirMarkerFilename)
				err = os.Remove(aufsMetaPath)
				if err != nil {
					return errors.Errorf("Error removing file at %s: %w", aufsMetaPath, err)
				}

			case MetaWhiteout:
				// create the 0x0 char device
				err := overlayutil.SetWhiteout(metaPath)
				if err != nil {
					return errors.Errorf("Error marking %s as whiteout: %w", metaPath, err)
				}
				metaDir, metaFile := filepath.Split(metaPath)
				aufsMetaPath := filepath.Join(metaDir, aufsutil.WhiteoutPrefix+metaFile)

				// chown the new char device with the old uid/gid
				uid, gid, err := osutil.GetUIDAndGID(aufsMetaPath)
				if err != nil {
					return errors.Errorf("Error getting UID and GUI for %s: %w", aufsMetaPath, err)
				}
				err = unix.Chown(metaPath, uid, gid)
				if err != nil {
					return errors.Errorf("Error chowning character device at %s: %w", metaPath, err)
				}

				err = os.Remove(aufsMetaPath)
				if err != nil {
					return errors.Errorf("Error removing file at %s: %w", aufsMetaPath, err)
				}

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
		logrus.Info("done")
	}

	logrus.Debug("moving from temporary root to overlay2 root")
	err = os.Rename(tempTargetRoot, overlayRoot)
	if err != nil {
		return errors.Errorf("Error moving %s to %s: %w", tempTargetRoot, overlayRoot, err)
	}

	logrus.Info("moving aufs images to overlay")
	aufsImageDir := filepath.Join(balenaEngineDir, "image", "aufs")
	overlayImageDir := filepath.Join(balenaEngineDir, "image", "overlay2")
	err = os.Rename(aufsImageDir, overlayImageDir)
	if err != nil {
		return errors.Errorf("Error moving %s to %s: %w", aufsImageDir, overlayImageDir, err)
	}

	logrus.Info("moving storage-driver of containers to overlay")
	containerDir := filepath.Join(balenaEngineDir, "containers")
	containerIDs, err := osutil.LoadIDs(containerDir)
	if err != nil {
		return errors.Errorf("Error listing containers in %s: %w", containerDir, err)
	}
	for _, containerID := range containerIDs {
		logrus := logrus.WithField("container_id", containerID)

		containerConfigPath := filepath.Join(containerDir, containerID, "config.v2.json")
		f, err := os.OpenFile(containerConfigPath, os.O_RDWR, 0600)
		if err != nil {
			return errors.Errorf("Error opening container config at %s: %w", containerConfigPath, err)
		}
		defer f.Close()

		var containerConfig = make(map[string]interface{})
		err = json.NewDecoder(f).Decode(&containerConfig)
		if err != nil {
			return errors.Errorf("Error parsing container config: %w", err)
		}
		containerConfig["Driver"] = "overlay2"
		err = json.NewEncoder(f).Encode(&containerConfig)
		if err != nil {
			return errors.Errorf("Error writing container config: %w", err)
		}

		logrus.Info("reconfigured storage-driver from aufs to overlay2")
	}

	logrus.Warn("daemon migration not done yet!")
	// sed -i "s/aufs/overlay2/g" /lib/systemd/system/balena.service
	// sed -i "s/aufs/overlay2/g" /etc/systemd/system/balena.service.d/balena.conf

	logrus.Info("finished migration")
	return nil
}
