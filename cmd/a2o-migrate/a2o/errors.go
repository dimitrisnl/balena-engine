package a2o // import "github.com/balena-os/balena-engine/cmd/a2o-migrate/a2o"

import errors "golang.org/x/xerrors"

var (
	ErrAuFSNotExists  = errors.New("aufs tree doesn't exists")
	ErrOverlay2Exists = errors.New("overlay2 tree exists")
)
