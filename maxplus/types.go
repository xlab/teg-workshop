package maxplus

import "fmt"

const (
	Inf  int = 2147483647
	_Inf int = -2147483647
)

var (
	Top = Gd{_Inf, Inf}
	Eps = Gd{Inf, _Inf}
	E   = Gd{0, 0}
)

type Gd struct {
	G, D int
}

type Poly []Gd

type Serie struct {
	P, Q Poly
	R    Gd
}

func (m *Gd) String() string {
	decor := func(prefix string, value int) string {
		if value == Inf {
			return prefix + "^inf"
		} else if value == _Inf {
			return prefix + "^-inf"
		} else if value == 1 {
			return prefix
		} else if value != 0 {
			return fmt.Sprintf("%s^%d", prefix, value)
		}
		return ""
	}
	switch {
	case m.isE():
		return "e"
	case m.isEps():
		return "eps"
	default:
		return decor("g", m.G) + decor("d", m.D)
	}
}

func (m *Gd) isE() bool {
	return m.G == 0 && m.D == 0
}

func (m *Gd) isEps() bool {
	return m.G == Eps.G && m.D == Eps.D
}

func (p Poly) isE() bool {
	if len(p) != 1 {
		return false
	}
	return p[0].isE()
}

func (p Poly) isEps() bool {
	if len(p) > 1 {
		return false
	} else if len(p) < 1 {
		return true
	}
	return p[0].isEps()
}

func (p Poly) String() (str string) {
	last := len(p) - 1
	for i, gd := range p {
		str += gd.String()
		if i < last {
			str += " + "
		}
	}
	return
}

func (s *Serie) String() (str string) {
	if !s.P.isEps() {
		str += s.P.String() + " + "
	}
	if len(s.Q) > 1 {
		str += "(" + s.Q.String() + ")"
	} else {
		str += s.Q.String()
	}
	if !s.R.isE() {
		str += "[" + s.R.String() + "]*"
	}
	return
}
