package a2o // import "github.com/docker/docker/cmd/a2o-migrate/a2o"

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"

	"github.com/docker/docker/cmd/a2o-migrate/aufsutil"
	"github.com/docker/docker/cmd/a2o-migrate/osutil"
	"github.com/docker/docker/cmd/a2o-migrate/overlayutil"
	"github.com/docker/docker/pkg/archive"
)

// Migrate migrates the state of the storage from aufs -> overlay2
func Migrate() error {
	logrus.Info("starting aufs -> overlay2 migration")

	// make sure we actually have an aufs tree to migrate from
	err := aufsutil.CheckRootExists(StorageRoot)
	if err != nil {
		return err
	}

	// make sure there isn't an overlay2 tree already
	err = overlayutil.CheckRootExists(StorageRoot)
	if err == nil {
		logrus.Warn("overlay root found, cleaning up...")
		err := os.Remove(overlayRoot())
		if err != nil {
			return fmt.Errorf("Error cleaning up overlay2 root: %v", err)
		}
	}

	var (
		state State
	)

	diffDir := filepath.Join(aufsRoot(), "diff")

	// Step 1:
	// Scan aufs layer data and build structure holding all the relevant information
	// needed to replicate on overlayfs.
	// We need to pay special attention to the whiteout metadata files used by aufs to
	// mark deleted files and empty directories.

	// get all layers
	layerIDs, err := osutil.LoadIDs(diffDir)
	if err != nil {
		return fmt.Errorf("Error loading layer ids: %v", err)
	}
	logrus.Debugf("layer ids in %s: %+#v", diffDir, layerIDs)

	for _, layerID := range layerIDs {
		logrus := logrus.WithField("layer_id", layerID)
		logrus.Debug("parsing layer")
		layer := Layer{ID: layerID}

		// get parent layers
		logrus.Debug("parsing parent ids")
		parentIDs, err := aufsutil.GetParentIDs(aufsRoot(), layerID)
		if err != nil {
			return fmt.Errorf("Error loading parent IDs for %s: %v", layerID, err)
		}
		layer.ParentIDs = parentIDs

		layerDir := filepath.Join(diffDir, layerID)
		logrus.Debug("parsing for metadata files")
		err = filepath.Walk(layerDir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			absPath, err := filepath.Rel(layerDir, path)
			if err != nil {
				return err
			}
			logrus := logrus.WithField("path", absPath)

			if aufsutil.IsWhiteout(fi.Name()) {
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
			return fmt.Errorf("Error walking filetree for %s: %v", layerID, err)
		}

		state.Layers = append(state.Layers, layer)
		logrus.Debug("done")
	}

	logrus.Infof("moving %d layer(s) to overlay", len(state.Layers))

	// Step 3:
	// Build up the overlayfs layer data from the state structure.
	// This ignores the special files at first and just replicates the data
	// (using hardlinks to save space).
	// In a second step we use the state.Meta data to delete aufs whiteout files
	// and create the special files / set file attributes used by overlayfs.

	// move to overlay filetree
	for _, layer := range state.Layers {
		logrus := logrus.WithField("layer_id", layer.ID)

		var (
			layerDir = filepath.Join(tempTargetRoot(), layer.ID)
		)

		logrus.Debugf("creating base dir %s", layerDir)
		// create /:layer_id dir
		err := os.MkdirAll(layerDir, os.ModeDir|0700)
		if err != nil {
			return fmt.Errorf("Error creating layer directory for %s: %v", layer.ID, err)
		}

		logrus.Debug("creating layer link")
		// create /:layer_id/link file and /l/:layer_ref file
		_, err = overlayutil.CreateLayerLink(tempTargetRoot(), layer.ID)
		if err != nil {
			return fmt.Errorf("Error creating layer link dir for %s: %v", layer.ID, err)
		}

		logrus.Debug("processing parent layers")
		// create /:layer_id/lower
		var lower string
		for _, parentID := range layer.ParentIDs {
			logrus := logrus.WithField("parent_layer_id", parentID)

			parentLayerDir := filepath.Join(tempTargetRoot(), parentID)
			ok, err := osutil.Exists(parentLayerDir, true)
			if err != nil {
				return fmt.Errorf("Error checking for parent layer dir for %s: %v", layer.ID, err)
			}
			if !ok {
				// parent layer hasn't been processed separately yet.
				logrus.Debugf("creating parent layer base dir %s", parentLayerDir)
				err := os.MkdirAll(parentLayerDir, os.ModeDir|0700)
				if err != nil {
					return fmt.Errorf("Error creating layer directory for parent layer %s: %v", parentID, err)
				}
			}
			logrus.Debug("creating parent layer link")
			parentRef, err := overlayutil.CreateLayerLink(tempTargetRoot(), parentID)
			if err != nil {
				return fmt.Errorf("Error creating layer link dir for parent layer %s: %v", parentID, err)
			}
			lower = overlayutil.AppendLower(lower, parentRef)
		}
		if lower != "" {
			lowerFile := filepath.Join(layerDir, "lower")
			logrus.Debugf("creating lower at %s", lowerFile)
			err := ioutil.WriteFile(lowerFile, []byte(lower), 0644)
			if err != nil {
				return fmt.Errorf("Error creating lower file for %s: %v", layer.ID, err)
			}
			layerWorkDir := filepath.Join(layerDir, "work")
			logrus.Debugf("creating work dir at %s", lowerFile)
			err = os.MkdirAll(layerWorkDir, os.ModeDir|0700)
			if err != nil {
				return fmt.Errorf("Error creating work dir for %s: %v", layer.ID, err)
			}
		}

		logrus.Debug("hardlinking aufs data to overlay")
		var (
			overlayLayerDir = filepath.Join(layerDir, "diff")
			aufsLayerDir    = filepath.Join(aufsRoot(), "diff", layer.ID)
		)
		err = replicate(aufsLayerDir, overlayLayerDir)
		if err != nil {
			return fmt.Errorf("Error moving layer data to overlay2: %v", err)
		}

		// migrate metadata files
		logrus.Debugf("processing %d metadata file(s)", len(layer.Meta))
		for _, meta := range layer.Meta {
			metaPath := filepath.Join(overlayLayerDir, meta.Path)

			switch meta.Type {
			case MetaOpaque:
				logrus.WithField("meta_type", "opaque").Debugf("translating %s to overlay", meta.Path)
				// set the opque xattr
				err := overlayutil.SetOpaque(metaPath)
				if err != nil {
					return fmt.Errorf("Error marking %s as opque: %v", metaPath, err)
				}
				// remove aufs metadata file
				aufsMetaFile := filepath.Join(metaPath, archive.WhiteoutOpaqueDir)
				err = os.Remove(aufsMetaFile)
				if err != nil {
					return fmt.Errorf("Error removing opque meta file: %v", err)
				}

			case MetaWhiteout:
				logrus.WithField("meta_type", "whiteout").Debugf("translating %s to overlay", meta.Path)
				// create the 0x0 char device
				err := overlayutil.SetWhiteout(metaPath)
				if err != nil {
					return fmt.Errorf("Error marking %s as whiteout: %v", metaPath, err)
				}
				metaDir, metaFile := filepath.Split(metaPath)
				aufsMetaFile := filepath.Join(metaDir, archive.WhiteoutPrefix+metaFile)

				// chown the new char device with the old uid/gid
				uid, gid, err := osutil.GetUIDAndGID(aufsMetaFile)
				if err != nil {
					return fmt.Errorf("Error getting UID and GID: %v", err)
				}
				err = unix.Chown(metaPath, uid, gid)
				if err != nil {
					return fmt.Errorf("Error chowning character device: %v", err)
				}

				err = os.Remove(aufsMetaFile)
				if err != nil {
					return fmt.Errorf("Error removing aufs whiteout file: %v", err)
				}

			case MetaOther:
				logrus.WithField("meta_type", "whiteoutmeta").Debugf("removing %s from overlay", meta.Path)
				err = os.Remove(metaPath)
				if err != nil {
					return fmt.Errorf("Error removing useless aufs meta file at: %v", err)
				}
			}
		}

		logrus.Debug("done")
	}

	// Step 4:
	// Finalize the migration:
	// - duplicate aufs images to $storageRoot/image/overlay2
	// - move temp dir holding overlay layer data to $storageRoot/overlay
	// - edit container config to use overlay storage driver

	logrus.Info("moving aufs images to overlay")
	var (
		aufsImageDir    = filepath.Join(StorageRoot, "image", "aufs")
		overlayImageDir = filepath.Join(StorageRoot, "image", "overlay2")
	)
	err = replicate(aufsImageDir, overlayImageDir)
	if err != nil {
		return fmt.Errorf("Error moving images from aufs to overlay: %v", err)
	}

	logrus.Info("moving layer data from temporary location to overlay2 root")
	err = os.Rename(tempTargetRoot(), overlayRoot())
	if err != nil {
		return fmt.Errorf("Error moving from temporary root: %v", err)
	}

	containerDir := filepath.Join(StorageRoot, "containers")
	containerIDs, err := osutil.LoadIDs(containerDir)
	if err != nil {
		return fmt.Errorf("Error listing containers: %v", err)
	}
	logrus.Infof("moving storage-driver of %d container(s) to overlay", len(containerIDs))
	for _, containerID := range containerIDs {
		err := switchContainerStorageDriver(containerID, "overlay2")
		if err != nil {
			return fmt.Errorf("Error rewriting container config for %s: %v", containerID, err)
		}
		logrus.WithField("container_id", containerID).Debug("reconfigured storage-driver from aufs to overlay2")
	}

	logrus.Info("finished migration")
	return nil
}
