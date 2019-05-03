package osutil // import "github.com/balena-os/balena-engine/cmd/a2o-migrate/osutil"

import (
	"os"
)

// Exists checks if a file  (or if isDir is set to "true" a directory) exists
func Exists(path string, isDir bool) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if fi.IsDir() != isDir {
		return false, nil
	}
	return true, nil
}
