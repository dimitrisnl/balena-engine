package a2o

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Cleanup should be run after the migration was successful.
// It removes the old aufs directory.
func Cleanup() error {
	logrus.Debug("starting cleanup")
	logrus.Warnf("This will remove %s", aufsRoot)

	err := os.Remove(aufsRoot)
	if err != nil {
		return err
	}

	return nil
}
