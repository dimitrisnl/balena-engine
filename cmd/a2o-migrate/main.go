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
	debug        = false
	printVersion = false
	runMigration = false
)

func main() {
	flag.BoolVar(&debug, "debug", debug, "enable debug logging")
	flag.BoolVar(&printVersion, "version", printVersion, "print version")
	flag.BoolVar(&runMigration, "migrate", runMigration, "migrate from aufs to overlay")
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	switch {
	case printVersion:
		fmt.Fprintf(os.Stdout, "a2o-migrate version %s (build %s)\n", GitVersion, BuildTime)
		os.Exit(0)

	case runMigration:
		if err := a2o.Migrate(); err != nil {
			logrus.Error(err)
			os.Exit(1)
		}


	default:
		flag.Usage()
	}
}
