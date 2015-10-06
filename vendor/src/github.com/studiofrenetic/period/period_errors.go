package period

import "errors"

var (
	OutOfRangeError     = errors.New("the submitted value is not contained within the valid range")
	ShouldOverlapsError = errors.New("Both Period objects should overlaps")
	BothShouldNotAbuts  = errors.New("Both object should not abuts")
)
