// Package dioid implements symbolic computations over dioid.
package dioid

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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

func scanGd(input string) (gd Gd, err error) {
	switch input {
	case "e":
		gd = Gd{0, 0}
		return
	case "gd":
		gd = Gd{1, 1}
		return
	default:
		rxG := regexp.MustCompile(`g\^(-?\d+)`)
		rxD := regexp.MustCompile(`d\^(-?\d+)`)
		strG := rxG.FindStringSubmatch(input)
		strD := rxD.FindStringSubmatch(input)
		if len(strG) < 2 && len(strD) < 2 {
			panic("should be at least one submatch")
		}
		if len(strG) > 0 {
			tmp, err := strconv.ParseInt(strG[1], 10, 32)
			if err != nil {
				return gd, err
			}
			gd.G = int(tmp)
		}
		if len(strD) > 0 {
			tmp, err := strconv.ParseInt(strD[1], 10, 32)
			if err != nil {
				return gd, err
			}
			gd.D = int(tmp)
		}
		return
	}
}

func (m Gd) String() string {
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
	case m.IsE():
		return "e"
	case m.IsEps():
		return "eps"
	default:
		return decor("g", m.G) + decor("d", m.D)
	}
}

func (m Gd) IsE() bool {
	return m.G == 0 && m.D == 0
}

func (m Gd) IsEps() bool {
	return m.G == Eps.G && m.D == Eps.D
}

func (p Poly) IsE() bool {
	if len(p) != 1 {
		return false
	}
	return p[0].IsE()
}

func (p Poly) IsEps() bool {
	if len(p) > 1 {
		return false
	} else if len(p) < 1 {
		return true
	}
	return p[0].IsEps()
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

func (s Serie) String() (str string) {
	if !s.P.IsEps() {
		str += s.P.String() + " + "
	}
	if len(s.Q) > 1 {
		str += "(" + s.Q.String() + ")"
	} else if !s.Q.IsE() {
		str += s.Q.String()
	}
	if !s.R.IsE() && !s.Q.IsE() {
		str += "x(" + s.R.String() + ")*"
	} else if !s.R.IsE() {
		str += "(" + s.R.String() + ")*"
	}
	return
}

func Latex(expr string) string {
	expr = strings.Replace(expr, "x", "", -1)
	expr = strings.Replace(expr, "+", "\\oplus", -1)
	expr = strings.Replace(expr, "*", "^{\\ast}", -1)
	expr = strings.Replace(expr, "eps", "\\varepsilon", -1)
	expr = strings.Replace(expr, "e", "e", -1)
	expr = strings.Replace(expr, "g", "\\gamma", -1)
	expr = strings.Replace(expr, "d", "\\delta", -1)
	rxPow := regexp.MustCompile(`\^(-?\d+)`)
	expr = rxPow.ReplaceAllString(expr, "^{${1}}")
	return base64.StdEncoding.EncodeToString([]byte(expr))
}

// utilitary
func (s Serie) RemoveGd(g, d int) Serie {
	for i, m := range s.P {
		if m.G == g && m.D == d {
			s.P = append(s.P[:i], s.P[i+1:]...)
			return SerieCanonize(s)
		}
	}
	for i, m := range s.Q {
		if m.G == g && m.D == d {
			s.Q = append(s.Q[:i], s.Q[i+1:]...)
			return SerieCanonize(s)
		}
	}
	return s
}
