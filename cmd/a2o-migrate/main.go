package main // import "github.com/balena-os/balena-engine/cmd/a2o-migrate"

import (
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/balena-os/balena-engine/cmd/a2o-migrate/a2o"
)

var ( // auto generated on build
	GitVersion = "undefined"
	BuildTime  = "undefined"
)

var ( // flag values
	debug             = false
	printVersion      = false
	modeAufsToOverlay = true
	modeOverlayToAufs = false
)

func main() {
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.BoolVar(&printVersion, "version", false, "print version")
	flag.BoolVar(&modeAufsToOverlay, "aufs-to-overlay", true, "migrate from aufs to overlay")
	flag.BoolVar(&modeOverlayToAufs, "overlay-to-aufs", false, "migrate back from overlay to aufs")
	flag.Parse()

	if printVersion {
		fmt.Fprintf(os.Stdout, "a2o-migrate version %s (build %s)\n", GitVersion, BuildTime)
		os.Exit(0)
	}

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	switch {
	case modeAufsToOverlay:
		if err := a2o.AuFSToOverlay(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
	case modeOverlayToAufs:
		fmt.Fprintf(os.Stderr, "error: not implemented!")
		os.Exit(1)
	}

	os.Exit(0)
}
