package maxplus

import (
	"testing"

	"github.com/remogatto/prettytest"
)

type testSuite struct {
	prettytest.Suite
}

func TestRunner(t *testing.T) {
	prettytest.RunWithFormatter(
		t,
		new(prettytest.TDDFormatter),
		new(testSuite),
	)
}

func (t *testSuite) TestPolySimply() {
	data := Poly{{0, 0}, {2, 0}, {2, 2}, {3, 3}}
	a := Poly{{0, 0}, {2, 2}, {3, 3}}
	b := PolySimply(data)
	t.Equal(a.String(), b.String())
}

func (t *testSuite) TestPolyStar() {
	data := Poly{{0, 0}, {2, 0}, {2, 2}, {3, 3}}
	a := Serie{
		P: Poly{{0, 0}},
		Q: Poly{{2, 2}},
		R: Gd{1, 1},
	}
	b := PolyStar(data)
	t.Equal(a.String(), b.String())
}

func BenchmarkPolySimply(b *testing.B) {
	data := Poly{
		{1, 2}, {1, 2}, {3, 3}, {3, 4},
		{4, 2}, {2, 2}, {5, 3}, {2, 3},
		{4, 6}, {1, 2}, {3, 7}, {3, 1},
	}
	for i := 0; i < b.N; i++ {
		_ = PolySimply(data)
	}
}

func BenchmarkPolyStar(b *testing.B) {
	data := Poly{
		{1, 2}, {1, 2}, {3, 3}, {3, 4},
		{4, 2}, {2, 2}, {5, 3}, {2, 3},
		{4, 6}, {1, 2}, {3, 7}, {3, 1},
	}
	for i := 0; i < b.N; i++ {
		_ = PolyStar(data)
	}
}
