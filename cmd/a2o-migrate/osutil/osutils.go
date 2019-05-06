package osutil // import "github.com/balena-os/balena-engine/cmd/a2o-migrate/osutil"

import (
	"os"
	"golang.org/x/sys/unix"
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

// GetUIDAndGID retrieves user and group id for path
func GetUIDAndGID(path string) (uid, gid int, err error) {
	var fi unix.Stat_t
	err = unix.Stat(path, &fi)
	if err != nil {
		return 0, 0, err
	}
	return int(fi.Uid), int(fi.Gid), nil
}
