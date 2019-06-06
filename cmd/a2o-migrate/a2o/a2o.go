package a2o

import "path/filepath"

const (
	storageRootDefault = "/var/lib/balena-engine"
)

var (
	StorageRoot = storageRootDefault

	aufsRoot       = filepath.Join(StorageRoot, "aufs")
	overlayRoot    = filepath.Join(StorageRoot, "overlay2")
	tempTargetRoot = filepath.Join(StorageRoot, "overlay2.temp")
)

// State models the state of the aufs/overlay2 directory
type State struct {
	Layers []Layer
}

// Layer is a single layer of an image
type Layer struct {
	ID        string
	ParentIDs []string
	Meta      []Meta
}

// MetaType describes the type of metadata
type MetaType int

const (
	// MetaOpaque means the directory should appear empty
	MetaOpaque = iota
	// MetaWhiteout means the file should not appear
	MetaWhiteout

	// MetaOther is a catch-all for everything else
	//
	// TODO(robertgzr): replace this with more meta types
	// for better mapping from a->o, o->a
	MetaOther
)

// Meta is extra data to make a layered FS work
// The information contained should enable reconstruction of metadata on both
// aufs and overlay
type Meta struct {
	Type MetaType
	// Path is the path to the affected file/dir
	Path string
}
