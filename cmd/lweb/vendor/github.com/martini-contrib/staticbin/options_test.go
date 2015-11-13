package staticbin

import "testing"

func TestRetrieveOptions(t *testing.T) {
	opt := Options{SkipLogging: true}
	retOpt := retrieveOptions([]Options{opt})
	if opt.SkipLogging != retOpt.SkipLogging {
		t.Errorf(
			"returned value is invalid [actual: %+v][expected: %+v]",
			opt,
			retOpt,
		)
	}
}
