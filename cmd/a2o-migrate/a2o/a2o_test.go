package a2o // import "github.com/balena-os/balena-engine/cmd/a2o-migrate/a2o"

import (
	"testing"

	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestAuFSToOverlay2(t *testing.T) {
	err := AuFSToOverlay2()
	assert.NilError(t, err)
}
