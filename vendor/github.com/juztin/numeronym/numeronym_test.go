package numeronym

import "testing"

type test struct {
	given, expected string
}

var tests = []test{
	{"a", "a"},
	{"_a", "_a"},
	{"a_", "a_"},
	{"_a_", "_a_"},
	{"ab", "ab"},
	{"_ab", "_ab"},
	{"ab_", "ab_"},
	{"_ab_", "_ab_"},
	{"abc", "a1c"},
	{"_abc", "_a1c"},
	{"abc_", "a1c_"},
	{"_abc_", "_a1c_"},
	{"abc_defgh", "a1c_d3h"},
	{"_abc_defgh", "_a1c_d3h"},
	{"abc_defgh_", "a1c_d3h_"},
	{"_abc_defgh_", "_a1c_d3h_"},
}

func TestParse(t *testing.T) {
	for _, test := range tests {
		b := Parse([]byte(test.given))
		if string(b) != test.expected {
			t.Errorf("Parse: expected %s; got %s", test.expected, string(b))
		}
	}
}

func benchmarkParse(bytes []byte, b *testing.B) {
	for n := 0; n < b.N; n++ {
		Parse(bytes)
	}
}

func BenchmarkParseExtraSmall(b *testing.B) { benchmarkParse([]byte("a"), b) }
func BenchmarkParseSmall(b *testing.B)      { benchmarkParse([]byte("abc"), b) }
func BenchmarkParseMedium(b *testing.B)     { benchmarkParse([]byte("abc-defgh"), b) }
func BenchmarkParseLarge(b *testing.B)      { benchmarkParse([]byte("abc-defgh-ijklmnop-qrstuv-wxy-z"), b) }
func BenchmarkParseExtraLarge(b *testing.B) {
	benchmarkParse([]byte("abc-defgh-ijklmnop-qrstuv-wxy-zabc-defgh-ijklmnop-qrstuv-wxy-z"), b)
}
