package staticbin

import "testing"

func TestClassicWithoutStatic(t *testing.T) {
	m := ClassicWithoutStatic()
	if m == nil {
		t.Errorf("returned value should not be nil.")
	}
}

func TestClassic(t *testing.T) {
	fnc := func(s string) ([]byte, error) {
		return []byte("test"), nil
	}
	m := Classic(fnc)
	if m == nil {
		t.Errorf("returned value should not be nil.")
	}
}
