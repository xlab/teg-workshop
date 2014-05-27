package dioid

// #cgo LDFLAGS: -lstdc++
// #include "bridge.h"
import "C"
import "unsafe"

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

func serie2ptr(s Serie) unsafe.Pointer {
	cp := poly2ptr(s.P)
	cq := poly2ptr(s.Q)
	cr := gd2ptr(s.R)
	cs := C.newSerie(cp, cq, cr)
	C.freePoly(cp)
	C.freePoly(cq)
	C.freeGd(cr)
	return cs
}

func ptr2serie(cserie unsafe.Pointer) Serie {
	p := ptr2poly(C.getP(cserie))
	q := ptr2poly(C.getQ(cserie))
	r := ptr2gd(C.getR(cserie))
	return Serie{p, q, r}
}

func PolySimply(p Poly) (result Poly) {
	cpoly := poly2ptr(p)
	C.simplyPoly(cpoly)
	result = ptr2poly(cpoly)
	C.freePoly(cpoly)
	return
}

func PolyStar(p Poly) (result Serie) {
	cpoly := poly2ptr(p)
	cserie := C.starPoly(cpoly)
	result = ptr2serie(cserie)
	C.freePoly(cpoly)
	C.freeSerie(cserie)
	return
}

func SerieStar(s Serie) (result Serie) {
	cserie := serie2ptr(s)
	cout := C.starSerie(cserie)
	result = ptr2serie(cout)
	C.freeSerie(cserie)
	C.freeSerie(cout)
	return
}

func PolyOplus(p1 Poly, p2 Poly) (result Poly) {
	cp1 := poly2ptr(p1)
	cp2 := poly2ptr(p2)
	result = ptr2poly(C.oplusPoly(cp1, cp2))
	C.freePoly(cp1)
	C.freePoly(cp2)
	return
}

func PolyOtimes(p1 Poly, p2 Poly) (result Poly) {
	cp1 := poly2ptr(p1)
	cp2 := poly2ptr(p2)
	result = ptr2poly(C.otimesPoly(cp1, cp2))
	C.freePoly(cp1)
	C.freePoly(cp2)
	return
}

func SerieOplus(s1 Serie, s2 Serie) (result Serie) {
	cs1 := serie2ptr(s1)
	cs2 := serie2ptr(s2)
	result = ptr2serie(C.oplusSerie(cs1, cs2))
	C.freeSerie(cs1)
	C.freeSerie(cs2)
	return
}

func SerieOtimes(s1 Serie, s2 Serie) (result Serie) {
	cs1 := serie2ptr(s1)
	cs2 := serie2ptr(s2)
	result = ptr2serie(C.otimesSerie(cs1, cs2))
	C.freeSerie(cs1)
	C.freeSerie(cs2)
	return
}

func SerieCanonize(s Serie) (result Serie) {
	cserie := serie2ptr(s)
	result = ptr2serie(cserie)
	C.freeSerie(cserie)
	return
}
