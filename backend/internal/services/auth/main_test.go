package auth_test

import (
	"flag"
	"os"
	"testing"

	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}
	os.Exit(testhelpers.Main(m))
}
