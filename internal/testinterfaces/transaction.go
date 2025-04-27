package testinterfaces

import "github.com/seanhuebl/unity-wealth/internal/testmodels"

type BaseAccess interface {
	BaseAccess() testmodels.BaseHTTPTestCase
}
