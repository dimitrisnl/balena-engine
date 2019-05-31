package osutil // import "github.com/docker/docker/cmd/a2o-migrate/osutil"

import (
	"io/ioutil"
	"os"
	"strings"

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

// Return all the directories
//
// from daemon/graphdriver/aufs/dirs.go
func LoadIDs(root string) ([]string, error) {
	dirs, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, d := range dirs {
		if d.IsDir() {
			out = append(out, d.Name())
		}
	}
	return out, nil
}

// Sed works like the sed(1) command on unix
func Sed(path, old, new string, n int) error {
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	newContents := strings.Replace(string(contents), old, new, n)
	f, err = os.OpenFile(path, os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = f.WriteString(newContents)
	if err != nil {
		return err
	}
	return nil
}
