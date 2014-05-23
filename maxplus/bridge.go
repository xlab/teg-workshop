package maxplus

// #cgo LDFLAGS: -lstdc++
// #include "bridge.h"
import "C"
import (
	"unsafe"
)

func gd2ptr(m Gd) unsafe.Pointer {
	return C.newGd(C.int(m.G), C.int(m.D))
}

func ptr2gd(cgd unsafe.Pointer) Gd {
	return Gd{
		int(C.getG(cgd)),
		int(C.getD(cgd)),
	}
}

func poly2ptr(p Poly) unsafe.Pointer {
	cpoly := C.newPoly()
	for _, m := range p {
		C.appendPoly(cpoly, C.int(m.G), C.int(m.D))
	}
	return cpoly
}

func ptr2poly(cpoly unsafe.Pointer) Poly {
	count := int(C.lenPoly(cpoly))
	poly := make([]Gd, count)
	for i := 0; i < count; i++ {
		m := C.getPoly(cpoly, C.int(i))
		poly[i] = Gd{
			int(C.getG(m)),
			int(C.getD(m)),
		}
	}
	return Poly(poly)
}

func serie2ptr(s *Serie) unsafe.Pointer {
	p := poly2ptr(s.P)
	q := poly2ptr(s.Q)
	r := gd2ptr(s.R)
	return C.newSerie(p, q, r)
}

func ptr2serie(cserie unsafe.Pointer) *Serie {
	p := ptr2poly(C.getP(cserie))
	q := ptr2poly(C.getQ(cserie))
	r := ptr2gd(C.getR(cserie))
	return &Serie{p, q, r}
}

func PolySimply(p Poly) Poly {
	cpoly := poly2ptr(p)
	C.simplyPoly(cpoly)
	poly := ptr2poly(cpoly)
	C.freePoly(cpoly)
	return poly
}

func PolyStar(p Poly) *Serie {
	cpoly := poly2ptr(p)
	cserie := C.starPoly(cpoly)
	serie := ptr2serie(cserie)
	C.freePoly(cpoly)
	C.freeSerie(cserie)
	return serie
}
