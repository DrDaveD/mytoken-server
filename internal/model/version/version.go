package version

import (
	"fmt"
)

// Version segments
const (
	MAJOR = 0
	MINOR = 3
	FIX   = 2
	DEV   = false
)

var version = fmt.Sprintf("%d.%d.%d", MAJOR, MINOR, FIX)
var devVersion = fmt.Sprintf("%s-dev", version)

// VERSION returns the current mytoken version
func VERSION() string {
	if DEV {
		return devVersion
	}
	return version
}
