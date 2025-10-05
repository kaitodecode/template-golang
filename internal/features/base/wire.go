package base

import (
	"github.com/google/wire"
)


var Set = wire.NewSet(
	NewBaseService,
)
