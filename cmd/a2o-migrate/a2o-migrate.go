package a2omigrate // import "github.com/docker/docker/cmd/a2o-migrate"

import (
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/docker/docker/cmd/a2o-migrate/a2o"
)

var ( // auto generated on build
	GitVersion = "undefined"
	BuildTime  = "undefined"
)

var ( // flag values
	debug          = false
	printVersion   = false
	runMigration   = false
	runCommit      = false
	runFailCleanup = false
)

func Main() {
	flag.BoolVar(&debug, "debug", debug, "enable debug logging")
	flag.BoolVar(&printVersion, "version", printVersion, "print version")
	flag.BoolVar(&runMigration, "migrate", runMigration, "migrate from aufs to overlay")
	flag.BoolVar(&runCommit, "commit", runCommit, "commit migration, removes aufs leftovers")
	flag.BoolVar(&runFailCleanup, "fail-cleanup", runFailCleanup, "recover from a failed migration")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [flags]\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "\nMigrate images, containers and daemon config files from aufs to overlay2...\n  while trying to not waste disk-space.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\nflags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nenvironment:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  BALENA_A2O_STORAGE_ROOT\n\tchange the storage root location (default: %s)\n", a2o.StorageRoot)
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
	}
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// parse env vars
	if env, ok := os.LookupEnv("BALENA_A2O_STORAGE_ROOT"); ok {
		a2o.StorageRoot = env
	}

	logrus.Warnf("storage root: %v", a2o.StorageRoot)

	switch {
	case printVersion:
		fmt.Fprintf(os.Stdout, "a2o-migrate version %s (build %s)\n", GitVersion, BuildTime)
		os.Exit(0)

	case runMigration:
		if err := a2o.Migrate(); err != nil {
			logrus.Error(err)
			os.Exit(1)
		}

	case runCommit:
		if err := a2o.Commit(); err != nil {
			logrus.Error(err)
			os.Exit(1)
		}

	case runFailCleanup:
		if err := a2o.FailCleanup(); err != nil {
			logrus.Error(err)
			os.Exit(1)
		}

	default:
		flag.Usage()
	}
}
