package a2o

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Cleanup should be run after the migration was successful.
// It removes the old aufs directory.
func Cleanup() error {
	logrus.Info("starting cleanup")

	err := removeDirIfExists(aufsRoot)
	if err != nil {
		return err
	}

	return nil
}
