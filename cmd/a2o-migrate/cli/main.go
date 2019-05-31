package main

import (
	"github.com/docker/docker/cmd/a2o-migrate"
)

var ( // auto generated on build
	GitVersion = "undefined"
	BuildTime  = "undefined"
)

func main() {
	a2omigrate.GitVersion = GitVersion
	a2omigrate.BuildTime = BuildTime
	a2omigrate.Main()
}
