#include "lib/gd.h"
#include "lib/poly.h"
#include "lib/serie.h"
#include "lib/smatrix.h"
#include "lib/tools.h"
#include "lib/vars.h"

#include "bridge.h"

gd_ *newGd(int g, int d) {
	return new gd;
}

poly_ *newPoly() {
	return new poly;
}

serie_ *newSerie(poly_ *p, poly_ *q, gd_ *r) {
	serie* s = new serie;
	s->init(*(poly*)p, *(poly*)q, *(gd*)r);
	return s;
}

void freeGd(gd_ *m) {
	delete (gd*)m;
}

void freePoly(poly_ *p) {
	delete (poly*)p;
}

void freeSerie(void* s) {
	delete (serie*)s;
}

void appendPoly(poly_ *p, int g, int d) {
	gd m(g, d);
	((poly*)p)->add(m);
}

void simplyPoly(poly_ *p) {
	((poly*)p)->simpli();
}

unsigned int lenPoly(poly_ *p) {
	return ((poly*)p)->getn();
}

gd_ *getPoly(poly_ *p, int n) {
	return &((poly*)p)->getpol(n);
}

int getG(gd_ *m) {
	return ((gd*)m)->getg();
}

int getD(gd_ *m) {
	return ((gd*)m)->getd();
}

poly_ *getP(serie_ *s) {
	return &((serie*)s)->getp();
}

poly_ *getQ(serie_ *s) {
	return &((serie*)s)->getq();
}

gd_ *getR(serie_ *s) {
	return &((serie*)s)->getr();
}

poly_ *oplusPoly(poly_ *p1, poly_ *p2) {
	return new poly(oplus(*(poly*)p1, *(poly*)p2));
}

poly_ *otimesPoly(poly_ *p1, poly_ *p2) {
	return new poly(otimes(*(poly*)p1, *(poly*)p2));
}

serie_ *oplusSerie(serie_ *s1, serie_ *s2) {
	return new serie(oplus(*(serie*)s1, *(serie*)s2));
}

serie_ *otimesSerie(serie_ *s1, serie_ *s2) {
	return new serie(otimes(*(serie*)s1, *(serie*)s2));
}

serie_ *starPoly(poly_ *p) {
	serie s(*(poly*)p);
	return new serie(star(s));
}

serie_ *starSerie(serie_ *s) {
	return new serie(star(*(serie*)s));
}
